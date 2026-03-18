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

## 5) Stop stack

```powershell
docker compose down
```

## 6) Stop and remove DB data

```powershell
docker compose down -v
```
