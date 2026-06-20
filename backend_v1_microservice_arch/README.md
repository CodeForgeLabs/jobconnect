# JobConnect — Backend (v1 microservice architecture)

Short README for local development and quick onboarding.

## Project overview

This repository contains the backend microservice architecture for JobConnect: a set of Go services communicating via gRPC/protobuf with an HTTP gateway, PostgreSQL databases, and object storage for user media. The frontend is a Next.js app in the `frontend/` folder.

## Tech stack

- Go (multiple services)
- gRPC / Protocol Buffers (`api/proto/`)
- Docker & Docker Compose
- PostgreSQL
- MinIO (object storage for avatars, portfolios, CVs)
- Next.js (frontend)

## Quick start (local)

From the repository root (`backend_v1_microservice_arch`) the usual development steps are:

1. Start the shared Postgres container:

```powershell
docker compose up -d postgres
```

2. Apply all DB migrations:

```powershell
.\scripts\migrate-all.ps1
# or on WSL / Linux: ./scripts/migrate-all.sh
```

If databases were previously migrated, baseline with:

```powershell
.\scripts\migrate-all.ps1 -BaselineExisting
```

3. Start backend services (example):

```powershell
docker compose up -d --build user auth job proposal contract wallet gateway
```

4. Check status and follow logs:

```powershell
docker compose ps
docker compose logs -f gateway auth user job proposal contract wallet
```

5. Stop the stack:

```powershell
docker compose down
```

See `RUNNING.md` for additional options and environment variable details.

## Important environment variables

- `.env` in this folder is used by services and compose.
- SMTP (auth service) — set `AUTH_SMTP_*` variables to enable OTP/emails.
- ReCAPTCHA & challenge (gateway) — set `GATEWAY_RECAPTCHA_*` and `GATEWAY_CHALLENGE_PROOF_*` as needed.
- User object storage (MinIO) — set `USER_AVATAR_STORAGE_*`, `USER_PORTFOLIO_STORAGE_*`, and `USER_CV_STORAGE_*` variables to configure buckets/endpoints.

See `RUNNING.md` for exact variable names and examples.

## Services (high level)

Key backend services live under `services/` and include (non-exhaustive):

- `auth` — authentication and OTP flows
- `user` — user profiles, avatars, portfolios, CVs
- `job` — job listings and search
- `proposal` — proposals and offers
- `contract` — contract lifecycle and milestones
- `payment`, `wallet` — payments and wallet handling
- `reviews`, `recommendation` — user reviews and recommendations

Each service contains a `cmd/` entrypoint and `internal/` implementation.

## Development notes

- Service contracts are defined in `api/proto/` and codegen is used for gRPC clients/servers.
- Use the Docker Compose setup for integration-style local testing.
- MinIO console is exposed on port `9001` by the compose stack (when enabled).

## Troubleshooting

- If migrations fail, try the `-BaselineExisting` flag or inspect `scripts/init` SQL.
- If object storage errors occur, verify MinIO endpoint and credentials in `.env`.
- For auth/email issues, verify SMTP settings are present and correct.

## Contributing

File a PR against this repo. Describe the change, which service it touches, and any migration or config updates required.

## Where to get more details

- Runbook and extended run instructions: `RUNNING.md`
- Service implementations: `services/*`
- Gateway routes and handlers: `gateway/internal/router` and `gateway/internal/handlers`

---
Generated and added by assistant — next: extract backend features for resume bullets.
