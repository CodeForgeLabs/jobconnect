# Backend Runbook

## 1) Start databases

Run from `backend`:

```powershell
docker compose up -d auth-db user-db job-db proposal-db contract-db wallet-db
```

## 2) Apply all migrations

```powershell
.\scripts\migrate-all.ps1
```

If databases were already migrated before using this script, baseline once:

```powershell
.\scripts\migrate-all.ps1 -BaselineExisting
```

Optional dry run:

```powershell
.\scripts\migrate-all.ps1 -DryRun
```

## 3) Start backend services

```powershell
docker compose up -d --build user auth job proposal contract wallet gateway
```

## 4) Check status and logs

```powershell
docker compose ps
docker compose logs -f gateway auth user job proposal contract wallet
```

## 4.1) Optional: Enable OTP emails via SMTP

Set these variables in `backend/.env` (used by `auth` service):

```powershell
AUTH_SMTP_HOST=smtp.example.com
AUTH_SMTP_PORT=587
AUTH_SMTP_TLS_MODE=starttls
AUTH_SMTP_USERNAME=mailer
AUTH_SMTP_PASSWORD=mailer-password
AUTH_SMTP_FROM_ADDRESS=no-reply@example.com
AUTH_SMTP_FROM_NAME=JobConnect
```

When `AUTH_SMTP_HOST` is empty, auth falls back to a no-op sender (no email delivery).

`AUTH_SMTP_TLS_MODE` options:
- `starttls` (default): port 587 style SMTP with STARTTLS
- `implicit`: SMTPS style connection (commonly port 465)
- `none`: plain SMTP (only for trusted local/dev SMTP)

## 5) Stop stack

```powershell
docker compose down
```

## 6) Stop and remove DB data

```powershell
docker compose down -v
```
