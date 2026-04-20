param(
    [switch]$DryRun,
    [switch]$BaselineExisting
)

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$backendDir = Split-Path -Parent $scriptDir
Set-Location $backendDir

$maps = @(
    @{ Name = "auth-db";     User = "auth";     Db = "jobconnect_auth";     Dir = "services/auth/migrations" },
    @{ Name = "user-db";     User = "user";     Db = "jobconnect_user";     Dir = "services/user/migrations" },
    @{ Name = "job-db";      User = "job";      Db = "jobconnect_job";      Dir = "services/job/migrations" },
    @{ Name = "proposal-db"; User = "proposal"; Db = "jobconnect_proposal"; Dir = "services/proposal/migrations" },
    @{ Name = "contract-db"; User = "contract"; Db = "jobconnect_contract"; Dir = "services/contract/migrations" },
    @{ Name = "wallet-db";   User = "wallet";   Db = "jobconnect_wallet";   Dir = "services/wallet/migrations" },
    @{ Name = "chat-db";     User = "chat";     Db = "jobconnect_chat";     Dir = "services/chat/migrations" },
    @{ Name = "connects-db"; User = "connects"; Db = "jobconnect_connects"; Dir = "services/connects/migrations" },
    @{ Name = "verification-db"; User = "verification"; Db = "jobconnect_verification"; Dir = "services/verification/migrations" },
    @{ Name = "reviews-db"; User = "reviews"; Db = "jobconnect_reviews"; Dir = "services/reviews/migrations" }
)

function Ensure-MigrationTable {
    param(
        [string]$Container,
        [string]$DbUser,
        [string]$DbName
    )

    docker compose exec -T $Container psql -U $DbUser -d $DbName -v ON_ERROR_STOP=1 -c "CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());" | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to ensure schema_migrations table for $Container"
    }
}

function Is-MigrationApplied {
    param(
        [string]$Container,
        [string]$DbUser,
        [string]$DbName,
        [string]$FileName
    )

    $escaped = $FileName.Replace("'", "''")
    $result = docker compose exec -T $Container psql -U $DbUser -d $DbName -tA -c "SELECT 1 FROM schema_migrations WHERE filename = '$escaped' LIMIT 1;"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed checking migration state for $Container on $FileName"
    }

    if (-not $result) {
        return $false
    }
    return ($result.Trim() -eq "1")
}

function Mark-MigrationApplied {
    param(
        [string]$Container,
        [string]$DbUser,
        [string]$DbName,
        [string]$FileName
    )

    $escaped = $FileName.Replace("'", "''")
    docker compose exec -T $Container psql -U $DbUser -d $DbName -v ON_ERROR_STOP=1 -c "INSERT INTO schema_migrations(filename) VALUES ('$escaped') ON CONFLICT (filename) DO NOTHING;" | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed recording migration state for $Container on $FileName"
    }
}

foreach ($m in $maps) {
    if (-not (Test-Path $m.Dir)) {
        throw "Migration directory not found: $($m.Dir)"
    }

    $files = Get-ChildItem $m.Dir -Filter "*.up.sql" | Sort-Object Name
    if ($files.Count -eq 0) {
        Write-Host "No .up.sql migrations found for $($m.Name)."
        continue
    }

    Ensure-MigrationTable -Container $m.Name -DbUser $m.User -DbName $m.Db

    foreach ($f in $files) {
        $msg = "Applying $($f.Name) to $($m.Name)"

        if ($BaselineExisting) {
            Write-Host "[BASELINE] Marking $($f.Name) as applied for $($m.Name)"
            Mark-MigrationApplied -Container $m.Name -DbUser $m.User -DbName $m.Db -FileName $f.Name
            continue
        }

        if (Is-MigrationApplied -Container $m.Name -DbUser $m.User -DbName $m.Db -FileName $f.Name) {
            Write-Host "[SKIP] $($f.Name) already applied on $($m.Name)"
            continue
        }

        if ($DryRun) {
            Write-Host "[DRY-RUN] $msg"
            continue
        }

        Write-Host $msg
        Get-Content $f.FullName -Raw | docker compose exec -T $m.Name psql -U $m.User -d $m.Db -v ON_ERROR_STOP=1
        if ($LASTEXITCODE -ne 0) {
            throw "Migration failed for $($m.Name) on file $($f.Name)"
        }

        Mark-MigrationApplied -Container $m.Name -DbUser $m.User -DbName $m.Db -FileName $f.Name
    }
}

if ($BaselineExisting) {
    Write-Host "Baseline complete."
} elseif ($DryRun) {
    Write-Host "Dry run complete."
} else {
    Write-Host "All migrations applied successfully."
}
