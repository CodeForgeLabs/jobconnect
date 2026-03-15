# Wallet Service

Wallet service provides ledger and escrow primitives for JobConnect.

## Scope

- Internal ledger (credit/debit in minor units).
- Hold lifecycle (place, release, capture).
- Balance and transaction history APIs.

No external payment provider is integrated in this phase.

## Environment

- `WALLET_GRPC_LISTEN_ADDR` default `:50059`
- `WALLET_POSTGRES_URL` required
- `AUTH_JWT_SECRET` required

## Run

1. Apply migrations from `migrations/`.
2. Generate protobuf stubs from backend root:
   - `buf generate --template buf.gen.wallet.yaml`
3. Start service from `services/wallet`:
   - `go mod tidy`
   - `go run ./cmd/walletd`

## Endpoint Roles

- Owner or internal role: `CreateWallet`, `GetWallet`, `GetBalance`, `ListTransactions`
- Internal role only (`system`, `admin`, `service`): `CreditWalletInternal`, `DebitWalletInternal`, `PlaceHold`, `ReleaseHold`, `CaptureHold`

## Phase 2 Integration Contract

Contract service should call wallet service with idempotency keys:

- milestone approved: `CaptureHold`
- contract canceled/disputed: `ReleaseHold`

Reference fields should include stable business IDs (e.g. `reference_type=milestone`, `reference_id=<milestone_id>`).
