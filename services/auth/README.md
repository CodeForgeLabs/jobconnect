## Auth service (gRPC)

### What’s included
- gRPC server skeleton with health + reflection
- Postgres connection pool
- Initial migrations for users/credentials/otp/sessions
- Proto definition at `api/proto/auth/v1/auth.proto`

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

