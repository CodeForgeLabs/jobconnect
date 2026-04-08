# Auth Service and Gateway API Documentation

## 1. Purpose and Audience

This document is the complete API reference for authentication flows in JobConnect.

It is written for:
- Frontend developers integrating login/session flows through HTTP.
- Backend integrators consuming internal auth contracts.

## 2. Architecture at a Glance

- Auth service exposes gRPC APIs in auth.v1 (internal service-to-service contract).
- Gateway exposes HTTP APIs under /api/v1/auth (frontend-facing facade over auth gRPC).
- Browser/mobile clients should call Gateway HTTP endpoints, not Auth gRPC directly.

## 3. Base URLs and Versioning

- Gateway base URL (local default): http://localhost:8080
- Gateway API prefix: /api/v1
- Auth proto package version: auth.v1

## 4. Global Conventions

### 4.1 Content Type

All JSON request bodies must use:

Content-Type: application/json

### 4.2 Access Token Header

For protected routes:

Authorization: Bearer <access_token>

### 4.3 Refresh Token Cookie

Gateway stores refresh tokens in an HTTP cookie.

Default cookie settings (env-configurable):
- Name: jc_refresh_token
- Path: /
- HttpOnly: true
- Secure: false in local defaults (set true in production)
- SameSite: lax

Implications for frontend:
- For browser calls, include credentials on requests that need refresh cookie.
- /auth/login and /auth/refresh update the refresh cookie.
- /auth/logout-everywhere clears the refresh cookie.

### 4.4 Date and Time Format

Session timestamps use RFC3339 UTC strings.

Example:
- 2026-03-20T16:42:51Z

### 4.5 Common Error Shape (Gateway)

Most error responses are:

{
  "error": "human readable message"
}

Rate-limit errors include additional fields:

{
  "error": "too many requests",
  "challenge_required": true,
  "challenge_endpoint": "/api/v1/auth/challenge"
}

### 4.6 gRPC to HTTP Status Mapping (Gateway)

When Gateway forwards auth gRPC errors, mapping is:

- InvalidArgument -> 400
- Unauthenticated -> 401
- PermissionDenied -> 403
- NotFound -> 404
- AlreadyExists -> 409
- Aborted -> 409
- FailedPrecondition -> 412
- ResourceExhausted -> 429
- Unimplemented -> 501
- Other -> 500

## 5. Security and Rate Limiting

### 5.1 Sensitive Auth Endpoint Limiter

The following endpoints are rate-limited by IP and challenge proof:
- POST /api/v1/auth/register
- POST /api/v1/auth/verify-email-otp
- POST /api/v1/auth/login
- POST /api/v1/auth/refresh
- POST /api/v1/auth/logout-everywhere
- POST /api/v1/auth/forgot-password
- POST /api/v1/auth/reset-password
- POST /api/v1/auth/email-change/request
- POST /api/v1/auth/email-change/confirm

Default limiter behavior:
- 1 request/second
- Burst: 5
- In-memory key TTL: 15 minutes

### 5.2 Challenge-Proof Bypass Flow

When rate-limited:
1. Client receives 429 with challenge_required=true.
2. Client calls POST /api/v1/auth/challenge with challenge_id and recaptcha_token.
3. On success, client receives challenge_proof.
4. Client retries original request with header X-Challenge-Proof: <challenge_proof>.

### 5.3 Protected Endpoints (Bearer Required)

- POST /api/v1/auth/email-change/request
- POST /api/v1/auth/email-change/confirm
- GET /api/v1/auth/sessions
- DELETE /api/v1/auth/sessions/:sessionId

## 6. Auth Service (gRPC) Reference

Service: AuthService

### 6.1 Register

RPC: Register(RegisterRequest) returns (RegisterResponse)

Request fields:
- email: string
- password: string
- first_name: string
- last_name: string
- role: string (client | freelancer | admin)
- accept_terms: bool

Response fields:
- user_id: string (UUID)
- otp_sent: bool

Notes:
- Creates account and triggers email OTP.
- Duplicate email typically returns AlreadyExists.

### 6.2 VerifyEmailOTP

RPC: VerifyEmailOTP(VerifyEmailOTPRequest) returns (VerifyEmailOTPResponse)

Request fields:
- email: string
- otp: string

Response fields:
- verified: bool

Notes:
- Invalid/expired OTP returns Unauthenticated.

### 6.3 Login

RPC: Login(LoginRequest) returns (LoginResponse)

Request fields:
- email: string
- password: string

Response fields:
- access_token: string
- refresh_token: string
- access_token_expires_in_seconds: int64

Notes:
- Invalid credentials return Unauthenticated.

### 6.4 Refresh

RPC: Refresh(RefreshRequest) returns (RefreshResponse)

Request fields:
- refresh_token: string

Response fields:
- access_token: string
- refresh_token: string
- access_token_expires_in_seconds: int64

Notes:
- Invalid/expired/revoked refresh token returns Unauthenticated.

### 6.5 LogoutEverywhere

RPC: LogoutEverywhere(LogoutEverywhereRequest) returns (LogoutEverywhereResponse)

Request fields:
- refresh_token: string

Response fields:
- ok: bool

Notes:
- Revokes current refresh-token session chain.

### 6.6 ForgotPassword

RPC: ForgotPassword(ForgotPasswordRequest) returns (ForgotPasswordResponse)

Request fields:
- email: string

Response fields:
- accepted: bool

Notes:
- Uses anti-enumeration behavior. accepted is true even if email is not registered.

### 6.7 ResetPassword

RPC: ResetPassword(ResetPasswordRequest) returns (ResetPasswordResponse)

Request fields:
- email: string
- otp: string
- new_password: string

Response fields:
- ok: bool

### 6.8 RequestEmailChange

RPC: RequestEmailChange(RequestEmailChangeRequest) returns (RequestEmailChangeResponse)

Request fields:
- user_id: string (UUID)
- new_email: string

Response fields:
- otp_sent: bool

### 6.9 ConfirmEmailChange

RPC: ConfirmEmailChange(ConfirmEmailChangeRequest) returns (ConfirmEmailChangeResponse)

Request fields:
- user_id: string (UUID)
- otp: string

Response fields:
- ok: bool

### 6.10 OAuthLogin

RPC: OAuthLogin(OAuthLoginRequest) returns (OAuthLoginResponse)

Request fields:
- provider: string
- provider_user_id: string
- email: string
- first_name: string
- last_name: string
- display_name: string
- role: string (optional for first-time signup)

Response fields:
- access_token: string
- refresh_token: string
- access_token_expires_in_seconds: int64
- is_new_user: bool

### 6.11 ListSessions

RPC: ListSessions(ListSessionsRequest) returns (ListSessionsResponse)

Request fields:
- user_id: string (UUID)

Response fields:
- sessions: repeated Session
  - session_id: string
  - created_at: string (RFC3339)
  - expires_at: string (RFC3339)
  - last_used_at: string (RFC3339 or empty)

### 6.12 RevokeSession

RPC: RevokeSession(RevokeSessionRequest) returns (RevokeSessionResponse)

Request fields:
- user_id: string (UUID)
- session_id: string (UUID)

Response fields:
- ok: bool

## 7. Gateway HTTP API Reference

Unless stated otherwise, all routes are under /api/v1.

### 7.1 Health

Method: GET
Path: /healthz
Auth: none
Rate limit: none

Success response (200):

{
  "status": "ok"
}

### 7.2 Register

Method: POST
Path: /api/v1/auth/register
Auth: none
Rate limit: sensitive limiter

Request body:

{
  "email": "dev@example.com",
  "password": "StrongPassword123!",
  "first_name": "Ari",
  "last_name": "Reyes",
  "role": "client",
  "accept_terms": true
}

Success response (200):

{
  "user_id": "c8f9f828-1f42-4f74-9227-277acde5480f",
  "otp_sent": true
}

Common errors:
- 400 invalid payload
- 409 email already exists
- 429 rate limited

### 7.3 Verify Email OTP

Method: POST
Path: /api/v1/auth/verify-email-otp
Auth: none
Rate limit: sensitive limiter

Request body:

{
  "email": "dev@example.com",
  "otp": "123456"
}

Success response (200):

{
  "verified": true
}

Common errors:
- 400 invalid payload
- 401 invalid or expired OTP
- 429 rate limited

### 7.4 Login

Method: POST
Path: /api/v1/auth/login
Auth: none
Rate limit: sensitive limiter

Request body:

{
  "email": "dev@example.com",
  "password": "StrongPassword123!"
}

Success response (200):

{
  "access_token": "<jwt>",
  "access_token_expires_in_seconds": 9000
}

Also sets cookie:
- Set-Cookie: jc_refresh_token=<token>; HttpOnly; Path=/; SameSite=Lax

Common errors:
- 400 invalid payload
- 401 invalid email or password
- 429 rate limited

### 7.5 Refresh

Method: POST
Path: /api/v1/auth/refresh
Auth: refresh cookie required
Rate limit: sensitive limiter

Required cookie:
- jc_refresh_token

Request body:
- none

Success response (200):

{
  "access_token": "<jwt>",
  "access_token_expires_in_seconds": 9000
}

Also rotates refresh cookie on success.

Common errors:
- 401 missing refresh cookie
- 401 invalid/expired refresh token
- 429 rate limited

### 7.6 Logout Everywhere

Method: POST
Path: /api/v1/auth/logout-everywhere
Auth: refresh cookie required
Rate limit: sensitive limiter

Required cookie:
- jc_refresh_token

Request body:
- none

Success response (200):

{
  "ok": true
}

Also clears refresh cookie.

Common errors:
- 401 missing refresh cookie
- 401 invalid refresh token
- 429 rate limited

### 7.7 Forgot Password

Method: POST
Path: /api/v1/auth/forgot-password
Auth: none
Rate limit: sensitive limiter

Request body:

{
  "email": "dev@example.com"
}

Success response (200):

{
  "accepted": true
}

Notes:
- accepted remains true for unknown email (anti-enumeration).

### 7.8 Reset Password

Method: POST
Path: /api/v1/auth/reset-password
Auth: none
Rate limit: sensitive limiter

Request body:

{
  "email": "dev@example.com",
  "otp": "123456",
  "new_password": "NewStrongPassword456!"
}

Success response (200):

{
  "ok": true
}

Common errors:
- 400 invalid payload
- 401 invalid/expired reset credentials
- 429 rate limited

### 7.9 Request Email Change

Method: POST
Path: /api/v1/auth/email-change/request
Auth: Bearer required
Rate limit: sensitive limiter

Headers:
- Authorization: Bearer <access_token>

Request body:

{
  "new_email": "new-email@example.com"
}

Success response (200):

{
  "otp_sent": true
}

Common errors:
- 401 missing or invalid authorization header
- 400 invalid payload
- 429 rate limited

### 7.10 Confirm Email Change

Method: POST
Path: /api/v1/auth/email-change/confirm
Auth: Bearer required
Rate limit: sensitive limiter

Headers:
- Authorization: Bearer <access_token>

Request body:

{
  "otp": "123456"
}

Success response (200):

{
  "ok": true
}

Common errors:
- 401 missing/invalid authorization header
- 401 invalid or expired OTP
- 429 rate limited

### 7.11 OAuth Start

Method: GET
Path: /api/v1/auth/oauth/:provider/start
Auth: none
Rate limit: none

Path params:
- provider: google | github

Query params:
- role (optional)

Behavior:
- Validates provider config.
- Issues signed oauth state.
- Redirects (302) to provider auth page.

Common errors:
- 400 unsupported provider or provider not configured
- 500 failed to issue oauth state

### 7.12 OAuth Callback

Method: GET
Path: /api/v1/auth/oauth/:provider/callback
Auth: none
Rate limit: none

Query params:
- state: required
- code: required

Success response (200):

{
  "access_token": "<jwt>",
  "access_token_expires_in_seconds": 9000,
  "is_new_user": true
}

Also sets refresh cookie.

Common errors:
- 400 missing or invalid oauth state/code
- 401 oauth exchange failed
- 400 provider unsupported/not configured

### 7.13 List Sessions

Method: GET
Path: /api/v1/auth/sessions
Auth: Bearer required
Rate limit: none

Headers:
- Authorization: Bearer <access_token>

Success response (200):

{
  "sessions": [
    {
      "session_id": "a07e88f0-445e-4130-b7e7-e8466c215398",
      "created_at": "2026-03-20T16:42:51Z",
      "expires_at": "2026-04-19T16:42:51Z",
      "last_used_at": "2026-03-20T17:00:08Z"
    }
  ]
}

Common errors:
- 401 missing/invalid authorization header

### 7.14 Revoke Session

Method: DELETE
Path: /api/v1/auth/sessions/:sessionId
Auth: Bearer required
Rate limit: none

Headers:
- Authorization: Bearer <access_token>

Path params:
- sessionId: UUID

Success response (200):

{
  "ok": true
}

Common errors:
- 401 missing/invalid authorization header
- 400 missing or malformed sessionId
- 403 forbidden session access
- 404 session not found

### 7.15 Challenge

Method: POST
Path: /api/v1/auth/challenge
Auth: none
Rate limit: none

Request body:

{
  "challenge_id": "login-20260320-abc",
  "recaptcha_token": "<token-from-recaptcha>"
}

Success response (200):

{
  "challenge_passed": true,
  "challenge_proof": "<signed-proof>",
  "challenge_id": "login-20260320-abc",
  "score": 0.9,
  "expires_at": "2026-03-20T16:55:45Z"
}

Common errors:
- 400 invalid payload
- 503 recaptcha not configured
- 401 challenge verification failed
- 401 challenge failed (low score or denied)

## 8. Postman-Style Flow Examples

### 8.1 Register + Verify OTP

Request:
POST /api/v1/auth/register

{
  "email": "frontend.user@example.com",
  "password": "Str0ngPassword!",
  "first_name": "Frontend",
  "last_name": "User",
  "role": "client",
  "accept_terms": true
}

Response 200:

{
  "user_id": "f6b5cb4a-e5a0-4187-940c-4a60f0f4db7f",
  "otp_sent": true
}

Request:
POST /api/v1/auth/verify-email-otp

{
  "email": "frontend.user@example.com",
  "otp": "123456"
}

Response 200:

{
  "verified": true
}

### 8.2 Login + Refresh + Protected Call

Request:
POST /api/v1/auth/login

{
  "email": "frontend.user@example.com",
  "password": "Str0ngPassword!"
}

Response 200:

{
  "access_token": "<jwt>",
  "access_token_expires_in_seconds": 9000
}

Request:
POST /api/v1/auth/refresh
Cookie: jc_refresh_token=<cookie value>

Response 200:

{
  "access_token": "<new jwt>",
  "access_token_expires_in_seconds": 9000
}

Request:
GET /api/v1/auth/sessions
Authorization: Bearer <access_token>

Response 200:

{
  "sessions": [
    {
      "session_id": "...",
      "created_at": "...",
      "expires_at": "...",
      "last_used_at": "..."
    }
  ]
}

### 8.3 Forgot + Reset Password

Request:
POST /api/v1/auth/forgot-password

{
  "email": "frontend.user@example.com"
}

Response 200:

{
  "accepted": true
}

Request:
POST /api/v1/auth/reset-password

{
  "email": "frontend.user@example.com",
  "otp": "123456",
  "new_password": "N3wPassword!"
}

Response 200:

{
  "ok": true
}

### 8.4 Rate Limit -> Challenge -> Retry

First request (rate-limited):
POST /api/v1/auth/login

Response 429:

{
  "error": "too many requests",
  "challenge_required": true,
  "challenge_endpoint": "/api/v1/auth/challenge"
}

Challenge request:
POST /api/v1/auth/challenge

{
  "challenge_id": "retry-login-1",
  "recaptcha_token": "<token>"
}

Challenge response 200:

{
  "challenge_passed": true,
  "challenge_proof": "<proof>",
  "challenge_id": "retry-login-1",
  "score": 0.9,
  "expires_at": "2026-03-20T17:01:00Z"
}

Retry request:
POST /api/v1/auth/login
X-Challenge-Proof: <proof>

### 8.5 OAuth Browser Flow

1. Browser opens GET /api/v1/auth/oauth/google/start?role=freelancer
2. Gateway redirects to provider auth page.
3. Provider redirects back to /api/v1/auth/oauth/google/callback?state=...&code=...
4. Gateway returns access token JSON and sets refresh cookie.

## 9. Frontend Integration Notes

- Keep access tokens in memory where possible; avoid long-term browser storage when not required.
- Always send Authorization: Bearer <access_token> for protected routes.
- Use credentialed requests when relying on refresh cookies.
- Implement automatic refresh-on-401 flow for token expiration.
- Handle challenge_required on 429 by calling /auth/challenge and retrying with X-Challenge-Proof.
- Consider clock skew when using access_token_expires_in_seconds and refresh slightly before expiry.

## 10. Backend Integrator Notes

- AuthService proto is the source of truth for internal contracts.
- Gateway intentionally hides refresh_token from HTTP body and uses cookie transport.
- If adding new auth RPCs, update both proto and gateway route docs in this file.

## 11. Completeness Checklist

Auth gRPC methods documented:
- Register
- VerifyEmailOTP
- Login
- Refresh
- LogoutEverywhere
- ForgotPassword
- ResetPassword
- RequestEmailChange
- ConfirmEmailChange
- OAuthLogin
- ListSessions
- RevokeSession

Gateway routes documented:
- GET /healthz
- POST /api/v1/auth/register
- POST /api/v1/auth/verify-email-otp
- POST /api/v1/auth/login
- POST /api/v1/auth/refresh
- POST /api/v1/auth/logout-everywhere
- POST /api/v1/auth/forgot-password
- POST /api/v1/auth/reset-password
- POST /api/v1/auth/email-change/request
- POST /api/v1/auth/email-change/confirm
- GET /api/v1/auth/oauth/:provider/start
- GET /api/v1/auth/oauth/:provider/callback
- GET /api/v1/auth/sessions
- DELETE /api/v1/auth/sessions/:sessionId
- POST /api/v1/auth/challenge
