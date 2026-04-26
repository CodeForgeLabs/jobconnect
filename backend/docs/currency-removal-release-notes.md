# Currency Removal and Hire Flow Fixes

## What changed
- Removed currency from job, contract, payment, and wallet service domains and APIs where it was part of the service-owned model.
- Added `contract_id` to the job hire response and propagated it through the gateway HTTP response.
- Kept `hired` as an internal orchestration state and blocked manual applicant-stage transitions from setting it.
- Replaced the contract service job-sync noop path with a real job client integration.
- Updated wallet semantics to owner-only identity instead of owner-plus-currency.
- Applied forward migrations to drop the removed currency columns and wallet uniqueness constraint.

## Database impact
- `jobs.currency` removed.
- `contracts.currency` removed.
- `payment_sessions.currency` removed.
- `wallet_accounts.currency` and `wallet_accounts_owner_currency_uniq` removed.

## Operator notes
- This is a breaking change for any client code that still sends or expects currency in the affected service payloads.
- Existing databases must run the new `0002_drop_currency` migrations for each affected service.
- The shared compose database was verified with the new migrations applied successfully.

## Verification
- `go test ./...` passed for `backend/services/job`.
- `go test ./...` passed for `backend/services/contract`.
- `go test ./...` passed for `backend/services/wallet`.
- `go test ./...` passed for `backend/services/payment`.
- `go test ./...` passed for `backend/gateway`.
- `./scripts/migrate-all.ps1` applied the new drop migrations successfully on the current compose stack.
