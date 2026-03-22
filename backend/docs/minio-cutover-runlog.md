# MinIO Avatar Cutover Run Log

## Step 1: Baseline before cutover
- Git branch: feat/avatar-object-storage
- Pending changes included by request:
  - backend/gateway/go.mod
  - backend/gateway/go.sum
  - backend/services/user/go.mod
  - backend/services/user/go.sum
  - backend/services/user/internal/application/avatar_test.go
  - backend/services/user/internal/config/config.go
- Runtime check: `docker compose ps` showed no `backend-minio-1` container running.
- DB schema check (`profile_avatars`): includes `content BYTEA`; no `storage_key`.
- Migration check (`schema_migrations` in `jobconnect_user`): up to `0004_freelancer_discovery_and_reputation.up.sql`.

## Step 2: Environment configuration
- Updated `backend/.env` (runtime file, git-ignored) with:
  - USER_AVATAR_STORAGE_PROVIDER=minio
  - USER_AVATAR_STORAGE_BUCKET=jobconnect-avatars
  - USER_AVATAR_STORAGE_ENDPOINT=minio:9000
  - USER_AVATAR_STORAGE_REGION=us-east-1
  - USER_AVATAR_STORAGE_ACCESS_KEY=minioadmin
  - USER_AVATAR_STORAGE_SECRET_KEY=minioadmin
  - USER_AVATAR_STORAGE_USE_SSL=false
  - USER_AVATAR_STORAGE_PATH_STYLE=true
  - USER_AVATAR_STORAGE_CREATE_BUCKET=true

## Step 3: Start MinIO and restart services
- Started MinIO with `docker compose up -d minio`.
- First `docker compose up -d --build user gateway` attempt failed due to transient Go proxy EOF errors while downloading modules for another service target during build.
- Docker engine temporarily became unavailable (`dockerDesktopLinuxEngine` pipe missing); recovered by restarting Docker Desktop engine.
- Restored full stack with `docker compose up -d`.
- Verified runtime state:
  - `backend-minio-1` healthy and mapped on `9000-9001`.
  - `backend-user-1` and `backend-gateway-1` running.
  - `user` logs show normal startup (`user gRPC listening on :50052`).

## Step 4: Apply migration 0005
- Pending.

## Step 5: Post-cutover verification
- Pending.
