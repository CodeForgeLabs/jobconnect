# User Service Profile Resource Endpoints

This document tracks portfolio, employment, education, certifications, and languages endpoints in the User Service.

## Gateway Routes

### Profile + Onboarding
- GET `/api/v1/users/me/profile` -> `GetMyProfile` + verification status composition (returns profile + completeness + verification)
- PATCH `/api/v1/users/me/profile` -> `PatchMyProfile` (single unified patch path for core + role fields)
- GET `/api/v1/users/me/onboarding-status` -> `GetMyOnboardingStatus` (returns `completeness`, `readiness`, and `steps`)

Profile patch ownership notes:
- `availability` and `hourly_rate` are updated through `PATCH /api/v1/users/me/profile` (the dedicated availability/rates endpoints were removed).
- `language` must be updated via account settings (`PATCH /api/v1/users/me/settings` with `ui_locale`).
- `tax_id` is patchable via `PATCH /api/v1/users/me/profile` only when verification is not `PENDING` and not `VERIFIED`.

### Account Settings
- GET `/api/v1/users/me/settings` -> `GetMySettings`
- PATCH `/api/v1/users/me/settings` -> `PatchMySettings` (`ui_locale` support)

### Portfolio
- POST `/api/v1/users/me/portfolio` -> `CreatePortfolioItem`
- GET `/api/v1/users/me/portfolio` -> `ListMyPortfolioItems`
- PATCH `/api/v1/users/me/portfolio/:itemId` -> `UpdatePortfolioItem`
- DELETE `/api/v1/users/me/portfolio/:itemId` -> `DeletePortfolioItem`
- PUT `/api/v1/users/me/portfolio:reorder` -> `ReorderPortfolioItems`
- GET `/api/v1/public/users/:userId/portfolio` -> `ListPublicPortfolioItems`

## Current Status

- Database schema has been added via migrations `0006` to `0008`.
- RPC contracts are defined in `api/proto/user/user.proto`.
- Gateway routes and handlers are wired for portfolio and public profile-resource reads.
- User service gRPC methods are implemented and use Postgres-backed repositories.
- Profile patching uses a single unified gateway-to-user-service call path.
- App locale updates are routed through account settings (`/users/me/settings`) instead of profile patch.
- Legacy duplicate availability/rates endpoints are removed from the gateway and user contract.

## Next Implementation Work

1. Add tests for public profile resource pagination behavior.
2. Add role/ownership authorization tests to verify public-only projections.
3. Add integration tests for gateway-to-gRPC error mapping.
