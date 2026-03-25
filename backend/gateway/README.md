# Gateway Service

## API Documentation

For complete authentication API documentation (Gateway HTTP and Auth gRPC contract), see:

- ../docs/auth-gateway-api.md

## Scope

Gateway provides frontend-facing HTTP endpoints under /api/v1 and forwards auth operations to the internal Auth gRPC service.

Core auth routes include:
- register, verify-email-otp, login, refresh, logout-everywhere
- forgot-password, reset-password
- email change request/confirm
- oauth start/callback
- session list/revoke
- challenge proof exchange for rate-limit bypass
