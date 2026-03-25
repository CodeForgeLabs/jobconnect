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
- Ran `./scripts/migrate-all.ps1` successfully.
- User DB migration `0005_avatar_object_storage_cutover.up.sql` applied.
- Migration output confirms hard cutover removed prior avatar rows (`DELETE 2`).
- Verified `schema_migrations` now includes `0005_avatar_object_storage_cutover.up.sql`.
- Verified `profile_avatars` schema now contains `storage_key` and no longer contains `content BYTEA`.

## Step 5: Post-cutover verification
- Performed authenticated upload against gateway endpoint `POST /api/v1/users/me/avatar` using a valid HS256 JWT for an existing profile user.
- Final upload succeeded with HTTP 200 and response metadata:
  - avatar_url `/profiles/9e61306a-a747-4f22-b97f-e081ddb6df5c/avatar`
  - content_type `image/png`
  - width `128`, height `128`, size_bytes `366`
- Verified `profile_avatars` row now stores metadata + key (no blob column):
  - storage_key `avatars/9e61306a-a747-4f22-b97f-e081ddb6df5c/current.png`
- Verified MinIO object exists on disk path:
  - `/data/jobconnect-avatars/avatars/9e61306a-a747-4f22-b97f-e081ddb6df5c/current.png`
- During verification, observed one transient mismatch where old `user` container build still referenced `content` column; resolved by rebuilding and restarting `user` service.

## Step 6: Final handoff summary
- Cutover status: complete.
- Evidence snapshot:
  - MinIO service healthy and running with ports `9000` (API) and `9001` (console).
  - User DB migration level includes `0005_avatar_object_storage_cutover.up.sql`.
  - `profile_avatars` schema uses `storage_key` and no longer has `content BYTEA`.
  - Gateway upload endpoint returned HTTP 200 and persisted avatar metadata.
  - MinIO object path exists for uploaded avatar under `/data/jobconnect-avatars/avatars/{user_id}/current.png`.
- Rollback caveat:
  - Down migration can reintroduce `content BYTEA` schema, but it cannot restore avatar rows deleted by hard-cutover migration `0005`.
