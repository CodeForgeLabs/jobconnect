# Contract RPC Live Test (Docker + One-Command Script)

## One-command usage
From repo root:

```powershell
powershell -ExecutionPolicy Bypass -File .\backend\scripts\contract-rpc-live-test.ps1
```

Optional flags:

```powershell
powershell -ExecutionPolicy Bypass -File .\backend\scripts\contract-rpc-live-test.ps1 -SkipComposeBuild
powershell -ExecutionPolicy Bypass -File .\backend\scripts\contract-rpc-live-test.ps1 -SkipMigrations
```

What it does:
1. Verifies tools (`docker`, `grpcurl`, `curl.exe`).
2. Starts backend stack with Docker Compose.
3. Runs migrations (`backend/scripts/migrate-all.ps1`) unless skipped.
4. Validates gateway/contract ports (`8080`, `50055`).
5. Logs in client + freelancer (register fallback if login fails).
6. Bootstraps fixture data (jobs, proposals, contracts, milestones, hourly logs).
7. Executes all `ContractService` RPCs using `backend/scripts/contract-rpc-cases.json`.
8. Runs extended negative scenarios (unauthenticated, wrong-role, internal-secret checks, invalid request, state violations).
9. Writes JSON summary to:
   - `backend/scripts/contract-rpc-live-test-summary.json`

## Required environment
Set in `backend/.env`:

```env
JOBCONNECT_INTERNAL_CALLER_SECRET=your-local-secret
```

The script fails fast if this value is missing because internal RPC checks require:
- `x-jobconnect-internal`
- `x-jobconnect-internal-secret`

## Case matrix interface
`backend/scripts/contract-rpc-cases.json` defines each RPC scenario with:
- `rpc`
- `actor` (`client`, `freelancer`, `internal:<service>`)
- `request`
- `expect_success_codes`
- optional: `extra_headers`, `wrong_role_actor`, `invalid_request`, `state_violation_request`, `save_from_response`

The runner resolves placeholders like `{{workflow_contract_id}}` from fixture/bootstrap context.

## Manual fallback (grpcurl / Postman)
If you need to run manually:

1. Start stack:
```powershell
cd backend
docker compose up --build -d
.\scripts\migrate-all.ps1
```

2. Bootstrap auth tokens:
- Use Postman collection `backend/docs/postman/proposal-gateway.postman_collection.json`
- Use env `backend/docs/postman/proposal-local.postman_environment.json`
- Run:
  - `Auth Bootstrap > Login Client`
  - `Auth Bootstrap > Login Freelancer`

3. Call contract gRPC methods:
- Address: `localhost:50055`
- Service: `contract.v1.ContractService`
- Public calls: add `authorization: Bearer <token>`
- Internal calls: add
  - `x-jobconnect-internal: <caller>`
  - `x-jobconnect-internal-secret: <secret>`
- `InternalGetJobOfferState` also needs client bearer auth.

## Troubleshooting
- **`JOBCONNECT_INTERNAL_CALLER_SECRET` missing**
  - Add it to `backend/.env`, then rerun.

- **Login fails for test users**
  - Script attempts register fallback automatically.
  - If still failing, inspect gateway/auth logs:
    ```powershell
    cd backend
    docker compose logs auth gateway --tail 200
    ```

- **RPC fails with `FailedPrecondition`**
  - This can be expected for specific negative/state-path checks.
  - Confirm expected codes in `contract-rpc-cases.json`.

- **Service unreachable (`localhost:50055` or `:8080`)**
  - Check container status:
    ```powershell
    cd backend
    docker compose ps
    docker compose logs contract gateway --tail 200
    ```
