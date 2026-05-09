import type { UserRole } from "@/features/login/loginSlice";

export interface ApiErrorPayload {
  error?: string;
  message?: string;
  challenge_required?: boolean;
  challenge_endpoint?: string;
}

export class ApiError extends Error {
  status: number;
  payload?: ApiErrorPayload;

  constructor(status: number, message: string, payload?: ApiErrorPayload) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.payload = payload;
  }
}

export interface AuthTokensResponse {
  access_token: string;
  access_token_expires_in_seconds: number;
}

export interface AuthSession {
  session_id: string;
  created_at: string;
  expires_at: string;
  last_used_at?: string;
}

export interface JwtPayload {
  sub?: string;
  role?: string;
  email?: string;
  first_name?: string;
  last_name?: string;
  exp?: number;
}

export function safeJsonParse<T>(value: string): T | null {
  try {
    return JSON.parse(value) as T;
  } catch {
    return null;
  }
}

export function decodeJwtPayload(token: string): JwtPayload | null {
  const parts = token.split(".");
  if (parts.length < 2) return null;

  try {
    const decoded = atob(parts[1].replace(/-/g, "+").replace(/_/g, "/"));
    return safeJsonParse<JwtPayload>(decoded);
  } catch {
    return null;
  }
}

export function normalizeRole(role?: string): UserRole {
  if (role === "client" || role === "freelancer" || role === "admin") {
    return role;
  }
  return "unknown";
}

export function expiresAtFromSeconds(ttlSeconds: number): number {
  return Date.now() + ttlSeconds * 1000;
}
