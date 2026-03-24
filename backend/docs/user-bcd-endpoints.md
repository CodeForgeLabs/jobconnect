# User Service B/C/D Endpoints

This document tracks the B/C/D rollout for User Service:
- B: Portfolio
- C: Employment + Education
- D: Certifications + Languages

## Gateway Routes

### Portfolio (B)
- POST `/api/v1/users/me/portfolio` -> `CreatePortfolioItem`
- GET `/api/v1/users/me/portfolio` -> `ListMyPortfolioItems`
- PATCH `/api/v1/users/me/portfolio/:itemId` -> `UpdatePortfolioItem`
- DELETE `/api/v1/users/me/portfolio/:itemId` -> `DeletePortfolioItem`
- PUT `/api/v1/users/me/portfolio:reorder` -> `ReorderPortfolioItems`
- GET `/api/v1/public/users/:userId/portfolio` -> `ListPublicPortfolioItems`

### Employment + Education (C)
- POST `/api/v1/users/me/employment` -> `CreateEmployment`
- GET `/api/v1/users/me/employment` -> `ListMyEmployment`
- PATCH `/api/v1/users/me/employment/:employmentId` -> `UpdateEmployment`
- DELETE `/api/v1/users/me/employment/:employmentId` -> `DeleteEmployment`
- GET `/api/v1/public/users/:userId/employment` -> `ListPublicEmployment`

- POST `/api/v1/users/me/education` -> `CreateEducation`
- GET `/api/v1/users/me/education` -> `ListMyEducation`
- PATCH `/api/v1/users/me/education/:educationId` -> `UpdateEducation`
- DELETE `/api/v1/users/me/education/:educationId` -> `DeleteEducation`
- GET `/api/v1/public/users/:userId/education` -> `ListPublicEducation`

### Certifications + Languages (D)
- POST `/api/v1/users/me/certifications` -> `CreateCertification`
- GET `/api/v1/users/me/certifications` -> `ListMyCertifications`
- PATCH `/api/v1/users/me/certifications/:certificationId` -> `UpdateCertification`
- DELETE `/api/v1/users/me/certifications/:certificationId` -> `DeleteCertification`
- GET `/api/v1/public/users/:userId/certifications` -> `ListPublicCertifications`

- PUT `/api/v1/users/me/languages` -> `UpsertLanguages`
- GET `/api/v1/users/me/languages` -> `GetMyLanguages`
- GET `/api/v1/public/users/:userId/languages` -> `GetPublicLanguages`

## Current Status

- Database schema for B/C/D has been added via migrations `0006` to `0008`.
- RPC contracts for B/C/D are defined in `api/proto/user/user.proto`.
- Gateway routes and handlers are wired.
- User service gRPC business logic for B/C/D RPCs is not implemented yet and currently returns gRPC unimplemented through embedded defaults.

## Next Implementation Work

1. Add domain models and repository ports for portfolio, employment, education, certifications, and languages.
2. Implement Postgres repositories in `services/user/internal/infrastructure/db`.
3. Add application use-cases in `services/user/internal/application`.
4. Wire use-cases and explicit RPC methods in `services/user/internal/adapter/grpc/user_server.go` and `cmd/userd/main.go`.
5. Add unit/integration tests for auth, ownership, and public projection behavior.
