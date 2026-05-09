"use client";

import { AUTH_ROUTES } from "@/lib/apiConfig";
import { apiClient } from "@/lib/apiClient";

export interface RegisterPayload {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  role: "client" | "freelancer";
  accept_terms: boolean;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface VerifyOtpPayload {
  email: string;
  otp: string;
}

export interface ForgotPasswordPayload {
  email: string;
}

export interface ResetPasswordPayload {
  email: string;
  otp: string;
  new_password: string;
}

export interface ChallengePayload {
  challenge_id: string;
  recaptcha_token: string;
}

export const authApi = {
  register: (payload: RegisterPayload, challengeProof?: string) =>
    apiClient.post<{ user_id: string; otp_sent: boolean }>(AUTH_ROUTES.register, payload, {
      challengeProof,
    }),

  verifyEmailOtp: (payload: VerifyOtpPayload, challengeProof?: string) =>
    apiClient.post<{ verified: boolean }>(AUTH_ROUTES.verifyEmailOtp, payload, {
      challengeProof,
    }),

  login: (payload: LoginPayload, challengeProof?: string) =>
    apiClient.post<{ access_token: string; access_token_expires_in_seconds: number }>(
      AUTH_ROUTES.login,
      payload,
      { challengeProof }
    ),

  refresh: () =>
    apiClient.post<{ access_token: string; access_token_expires_in_seconds: number }>(
      AUTH_ROUTES.refresh
    ),

  forgotPassword: (payload: ForgotPasswordPayload, challengeProof?: string) =>
    apiClient.post<{ accepted: boolean }>(AUTH_ROUTES.forgotPassword, payload, {
      challengeProof,
    }),

  resetPassword: (payload: ResetPasswordPayload, challengeProof?: string) =>
    apiClient.post<{ ok: boolean }>(AUTH_ROUTES.resetPassword, payload, {
      challengeProof,
    }),

  logoutEverywhere: () =>
    apiClient.post<{ ok: boolean }>(AUTH_ROUTES.logoutEverywhere, undefined),

  listSessions: () =>
    apiClient.get<{ sessions: Array<{ session_id: string; created_at: string; expires_at: string; last_used_at?: string }> }>(
      AUTH_ROUTES.sessions,
      { auth: true }
    ),

  revokeSession: (sessionId: string) =>
    apiClient.delete<{ ok: boolean }>(`${AUTH_ROUTES.sessions}/${sessionId}`, {
      auth: true,
    }),

  requestEmailChange: (newEmail: string, challengeProof?: string) =>
    apiClient.post<{ otp_sent: boolean }>(
      AUTH_ROUTES.emailChangeRequest,
      { new_email: newEmail },
      { auth: true, challengeProof }
    ),

  confirmEmailChange: (otp: string, challengeProof?: string) =>
    apiClient.post<{ ok: boolean }>(
      AUTH_ROUTES.emailChangeConfirm,
      { otp },
      { auth: true, challengeProof }
    ),

  solveChallenge: (payload: ChallengePayload) =>
    apiClient.post<{ challenge_passed: boolean; challenge_proof: string; challenge_id: string }>(
      AUTH_ROUTES.challenge,
      payload
    ),
};
