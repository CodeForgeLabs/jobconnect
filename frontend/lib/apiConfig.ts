export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL?.replace(/\/$/, "") ||
  "";

export const AUTH_ROUTES = {
  register: "/api/v1/auth/register",
  verifyEmailOtp: "/api/v1/auth/verify-email-otp",
  login: "/api/v1/auth/login",
  refresh: "/api/v1/auth/refresh",
  logoutEverywhere: "/api/v1/auth/logout-everywhere",
  forgotPassword: "/api/v1/auth/forgot-password",
  resetPassword: "/api/v1/auth/reset-password",
  emailChangeRequest: "/api/v1/auth/email-change/request",
  emailChangeConfirm: "/api/v1/auth/email-change/confirm",
  sessions: "/api/v1/auth/sessions",
  challenge: "/api/v1/auth/challenge",
  oauthStart: (provider: string) => `/api/v1/auth/oauth/${provider}/start`,
} as const;

export const USER_ROUTES = {
  profile: "/api/v1/users/me/profile",
  onboardingStatus: "/api/v1/users/me/onboarding-status",
  settings: "/api/v1/users/me/settings",
  avatarUploadUrl: "/api/v1/users/me/avatar/upload-url",
  avatar: "/api/v1/users/me/avatar",
  cvUploadUrl: "/api/v1/users/me/cv/upload-url",
  cv: "/api/v1/users/me/cv",
  portfolioUploadUrl: "/api/v1/users/me/portfolio/media/upload-url",
  portfolio: "/api/v1/users/me/portfolio",
  workPreferences: "/api/v1/users/me/work-preferences",
  hiringPreferences: "/api/v1/users/me/hiring-preferences",
  savedFreelancers: "/api/v1/users/me/saved-freelancers",
  freelancerNotes: "/api/v1/users/me/freelancer-notes",
} as const;

export const STORAGE_HEADERS = {
  challengeProof: "X-Challenge-Proof",
};
