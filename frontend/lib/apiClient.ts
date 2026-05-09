"use client";

import { store } from "@/store/store";
import {
  clearAuthState,
  selectAccessToken,
  selectAccessTokenExpiresAt,
  setAuthState,
  updateAccessToken,
} from "@/features/login/loginSlice";
import { API_BASE_URL, AUTH_ROUTES, STORAGE_HEADERS } from "@/lib/apiConfig";
import {
  ApiError,
  decodeJwtPayload,
  expiresAtFromSeconds,
  normalizeRole,
  type ApiErrorPayload,
  type AuthTokensResponse,
} from "@/lib/apiTypes";

export interface RequestOptions extends RequestInit {
  auth?: boolean;
  retryOnUnauthorized?: boolean;
  challengeProof?: string;
}

let refreshInFlight: Promise<string> | null = null;

async function parseErrorPayload(response: Response): Promise<ApiErrorPayload | undefined> {
  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.includes("application/json")) {
    return undefined;
  }

  try {
    return (await response.json()) as ApiErrorPayload;
  } catch {
    return undefined;
  }
}

function getAuthHeader(): string | null {
  const state = store.getState();
  return selectAccessToken(state) ?? null;
}

async function refreshAccessToken(): Promise<string> {
  if (refreshInFlight) {
    return refreshInFlight;
  }

  refreshInFlight = (async () => {
    const response = await fetch(`${API_BASE_URL}${AUTH_ROUTES.refresh}`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
    });

    if (!response.ok) {
      store.dispatch(clearAuthState());
      const payload = await parseErrorPayload(response);
      throw new ApiError(response.status, payload?.error || "Failed to refresh session", payload);
    }

    const data = (await response.json()) as AuthTokensResponse;
    const payload = decodeJwtPayload(data.access_token);

    store.dispatch(
      updateAccessToken({
        accessToken: data.access_token,
        expiresAt: expiresAtFromSeconds(data.access_token_expires_in_seconds),
      })
    );

    // Refresh can return a new role claim, keep it in sync.
    if (payload?.role) {
      store.dispatch(
        setAuthState({
          accessToken: data.access_token,
          expiresAt: expiresAtFromSeconds(data.access_token_expires_in_seconds),
          userRole: normalizeRole(payload.role),
          user: {
            userId: payload.sub,
            email: payload.email,
            firstName: payload.first_name,
            lastName: payload.last_name,
          },
        })
      );
    }

    return data.access_token;
  })();

  try {
    return await refreshInFlight;
  } finally {
    refreshInFlight = null;
  }
}

async function send<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const {
    auth = false,
    retryOnUnauthorized = true,
    headers,
    challengeProof,
    ...rest
  } = options;

  const composedHeaders = new Headers(headers ?? {});

  if (!composedHeaders.has("Content-Type") && rest.body && !(rest.body instanceof FormData)) {
    composedHeaders.set("Content-Type", "application/json");
  }

  if (auth) {
    const token = getAuthHeader();
    if (token) {
      composedHeaders.set("Authorization", `Bearer ${token}`);
    }
  }

  if (challengeProof) {
    composedHeaders.set(STORAGE_HEADERS.challengeProof, challengeProof);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...rest,
    headers: composedHeaders,
    credentials: "include",
  });

  if (response.ok) {
    if (response.status === 204) {
      return undefined as T;
    }

    const contentType = response.headers.get("content-type") ?? "";
    if (!contentType.includes("application/json")) {
      return undefined as T;
    }

    return (await response.json()) as T;
  }

  if (response.status === 401 && auth && retryOnUnauthorized) {
    const refreshedToken = await refreshAccessToken();
    return send<T>(path, {
      ...options,
      retryOnUnauthorized: false,
      headers: {
        ...(headers ?? {}),
        Authorization: `Bearer ${refreshedToken}`,
      },
    });
  }

  const payload = await parseErrorPayload(response);
  throw new ApiError(
    response.status,
    payload?.error || payload?.message || `Request failed (${response.status})`,
    payload
  );
}

export const apiClient = {
  get: <T>(path: string, options: RequestOptions = {}) =>
    send<T>(path, { ...options, method: "GET" }),
  post: <T>(path: string, body?: unknown, options: RequestOptions = {}) =>
    send<T>(path, {
      ...options,
      method: "POST",
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  patch: <T>(path: string, body?: unknown, options: RequestOptions = {}) =>
    send<T>(path, {
      ...options,
      method: "PATCH",
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  put: <T>(path: string, body?: unknown, options: RequestOptions = {}) =>
    send<T>(path, {
      ...options,
      method: "PUT",
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  delete: <T>(path: string, options: RequestOptions = {}) =>
    send<T>(path, { ...options, method: "DELETE" }),
};

export async function ensureFreshAuthToken(): Promise<void> {
  const state = store.getState();
  const expiresAt = selectAccessTokenExpiresAt(state);
  if (!expiresAt) return;

  const skewMs = 30 * 1000;
  if (Date.now() + skewMs >= expiresAt) {
    await refreshAccessToken();
  }
}
