#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
BASELINE_EXISTING=false

for arg in "$@"; do
    case "$arg" in
        --dry-run)         DRY_RUN=true ;;
        --baseline-existing) BASELINE_EXISTING=true ;;
        *) echo "Unknown argument: $arg"; exit 1 ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"
cd "$BACKEND_DIR"

declare -a NAMES=( "postgres" "postgres" "postgres" "postgres" "postgres" "postgres" "postgres" "postgres" "postgres" )
declare -a USERS=( "auth"        "user"        "job"         "proposal"        "contract"        "wallet"        "chat"        "connects"         "verification"        )
declare -a DBS=(   "jobconnect_auth" "jobconnect_user" "jobconnect_job" "jobconnect_proposal" "jobconnect_contract" "jobconnect_wallet" "jobconnect_chat" "jobconnect_connects" "jobconnect_verification" )
declare -a DIRS=(
    "services/auth/migrations"
    "services/user/migrations"
    "services/job/migrations"
    "services/proposal/migrations"
    "services/contract/migrations"
    "services/wallet/migrations"
    "services/chat/migrations"
    "services/connects/migrations"
    "services/verification/migrations"
)

ensure_migration_table() {
    local container="$1" db_user="$2" db_name="$3"
    docker compose exec -T "$container" psql -U "$db_user" -d "$db_name" -v ON_ERROR_STOP=1 -c \
        "CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());" > /dev/null
}

is_migration_applied() {
    local container="$1" db_user="$2" db_name="$3" filename="$4"
    local escaped="${filename//"'"/"''"}"
    local result
    result=$(docker compose exec -T "$container" psql -U "$db_user" -d "$db_name" -tA -c \
        "SELECT 1 FROM schema_migrations WHERE filename = '$escaped' LIMIT 1;")
    [[ "${result// /}" == "1" ]]
}

mark_migration_applied() {
    local container="$1" db_user="$2" db_name="$3" filename="$4"
    local escaped="${filename//"'"/"''"}"
    docker compose exec -T "$container" psql -U "$db_user" -d "$db_name" -v ON_ERROR_STOP=1 -c \
        "INSERT INTO schema_migrations(filename) VALUES ('$escaped') ON CONFLICT (filename) DO NOTHING;" > /dev/null
}

for i in "${!NAMES[@]}"; do
    name="${NAMES[$i]}"
    db_user="${USERS[$i]}"
    db_name="${DBS[$i]}"
    dir="${DIRS[$i]}"

    if [[ ! -d "$dir" ]]; then
        echo "Migration directory not found: $dir"
        exit 1
    fi

    mapfile -t files < <(find "$dir" -maxdepth 1 -name "*.up.sql" | sort)

    if [[ ${#files[@]} -eq 0 ]]; then
        echo "No .up.sql migrations found for $name."
        continue
    fi

    ensure_migration_table "$name" "$db_user" "$db_name"

    for f in "${files[@]}"; do
        filename="$(basename "$f")"

        if [[ "$BASELINE_EXISTING" == true ]]; then
            echo "[BASELINE] Marking $filename as applied for $name"
            mark_migration_applied "$name" "$db_user" "$db_name" "$filename"
            continue
        fi

        if is_migration_applied "$name" "$db_user" "$db_name" "$filename"; then
            echo "[SKIP] $filename already applied on $name"
            continue
        fi

        if [[ "$DRY_RUN" == true ]]; then
            echo "[DRY-RUN] Applying $filename to $name"
            continue
        fi

        echo "Applying $filename to $name"
        docker compose exec -T "$name" psql -U "$db_user" -d "$db_name" -v ON_ERROR_STOP=1 < "$f"
        mark_migration_applied "$name" "$db_user" "$db_name" "$filename"
    done
done

if [[ "$BASELINE_EXISTING" == true ]]; then
    echo "Baseline complete."
elif [[ "$DRY_RUN" == true ]]; then
    echo "Dry run complete."
else
    echo "All migrations applied successfully."
fi
