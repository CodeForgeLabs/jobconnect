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
- Pending.

## Step 3: Start MinIO and restart services
- Pending.

## Step 4: Apply migration 0005
- Pending.

## Step 5: Post-cutover verification
- Pending.
