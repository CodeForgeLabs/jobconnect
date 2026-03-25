## Auth service (gRPC)

## API Documentation

For the complete authentication API reference (Gateway HTTP endpoints + Auth gRPC contract), see:

- `../../docs/auth-gateway-api.md`

### What’s included
- gRPC server skeleton with health + reflection
- Postgres connection pool
- Initial migrations for users/credentials/otp/sessions
- Proto definition at `api/proto/auth/v1/auth.proto`

### Regenerate Go from protos (from repo root)
Requires `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` on PATH. From `C:\JobConnect`:

```powershell
# Ensure Go plugin bin is on PATH (e.g. $env:USERPROFILE\go\bin)
protoc --proto_path=api/proto --go_out=services/auth/gen --go_opt=paths=source_relative --go-grpc_out=services/auth/gen --go-grpc_opt=paths=source_relative api/proto/auth/v1/auth.proto
```

Alternatively use Buf: add `buf.yaml` and `buf.gen.yaml` at repo root, then `buf generate`.

### Run locally
1. Start Postgres and create a database (e.g. `auth_db`)
2. Apply migrations from `services/auth/migrations/`
3. Run:

```powershell
$env:AUTH_POSTGRES_URL="postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"
$env:AUTH_GRPC_LISTEN_ADDR=":50051"
$env:GOTOOLCHAIN="local"
go run ./cmd/authd
```

