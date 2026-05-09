"use client";

import { setAuthState, type UserRole } from "@/features/login/loginSlice";
import { store } from "@/store/store";
import { decodeJwtPayload, expiresAtFromSeconds, normalizeRole } from "@/lib/apiTypes";

export function applyLoginToken(accessToken: string, expiresInSeconds: number) {
  const payload = decodeJwtPayload(accessToken);
  const role: UserRole = normalizeRole(payload?.role);

  store.dispatch(
    setAuthState({
      accessToken,
      expiresAt: expiresAtFromSeconds(expiresInSeconds),
      userRole: role,
      user: {
        userId: payload?.sub,
        email: payload?.email,
        firstName: payload?.first_name,
        lastName: payload?.last_name,
      },
    })
  );
}

export function redirectPathForRole(role: UserRole) {
  if (role === "freelancer") return "/freelancer/dashboard";
  if (role === "client") return "/client/dashboard";
  return "/account";
}
