export interface RegisterRequestPayload {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  role: "client" | "freelancer";
  accept_terms: boolean;
}

export interface RegisterResponsePayload {
  user_id: string;
  otp_sent: boolean;
}

export interface VerifyEmailOtpRequestPayload {
  email: string;
  otp: string;
}

export interface VerifyEmailOtpResponsePayload {
  verified: boolean;
}

export interface LoginRequestPayload {
  email: string;
  password: string;
}

export interface AuthTokensPayload {
  access_token: string;
  access_token_expires_in_seconds: number;
}

export interface ForgotPasswordRequestPayload {
  email: string;
}

export interface ForgotPasswordResponsePayload {
  accepted: boolean;
}

export interface ResetPasswordRequestPayload {
  email: string;
  otp: string;
  new_password: string;
}

export interface ChangePasswordRequestPayload {
  current_password: string;
  new_password: string;
}

export interface GenericOkResponse {
  ok: boolean;
}

export interface RequestEmailChangePayload {
  new_email: string;
}

export interface RequestEmailChangeResponsePayload {
  otp_sent: boolean;
}

export interface ConfirmEmailChangePayload {
  otp: string;
}

export interface ChallengePayload {
  challenge_id: string;
  recaptcha_token: string;
}

export interface ChallengeResponsePayload {
  challenge_passed: boolean;
  challenge_proof: string;
  challenge_id: string;
}

export interface AuthSessionItemPayload {
  session_id: string;
  created_at: string;
  expires_at: string;
  last_used_at?: string;
}

export interface ListSessionsResponsePayload {
  sessions: AuthSessionItemPayload[];
}

