#!/bin/bash
# Database Migration Rollback Script for FunctionFly
# Usage: ./scripts/rollback-migration.sh [version|down|status]
#
# Examples:
#   ./scripts/rollback-migration.sh status           # Show current migration status
#   ./scripts/rollback-migration.sh down              # Roll back the last migration
#   ./scripts/rollback-migration.sh 20250101000000    # Roll back to specific version
#   DRY_RUN=1 ./scripts/rollback-migration.sh down    # Preview without applying
#
# Prerequisites:
#   - Database connection via DATABASE_URL or DB_HOST/DB_USER/DB_PASSWORD/DB_NAME
#   - Migration files in migrations/ directory with .up.sql and .down.sql
#
# Warning: Always test rollbacks in staging before production!

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MIGRATIONS_DIR="${PROJECT_ROOT}/migrations"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

DRY_RUN="${DRY_RUN:-0}"
VERBOSE="${VERBOSE:-0}"

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment if .env exists
if [ -f "${PROJECT_ROOT}/.env" ]; then
    set -a
    source "${PROJECT_ROOT}/.env"
    set +a
fi

show_usage() {
    cat <<EOF
Usage: $0 [command] [options]

Commands:
    status              Show current migration status
    down                Roll back the last applied migration
    version             Roll back to a specific version (requires version argument)
    list                List all migrations (applied and pending)

Options:
    --dry-run           Preview changes without applying
    --verbose           Show detailed SQL being executed

Examples:
    $0 status
    $0 down
    DRY_RUN=1 $0 down
    $0 version 20250101000000

Environment:
    DATABASE_URL         Full database connection string
    DB_HOST              Database host (default: localhost)
    DB_PORT              Database port (default: 5432)
    DB_USER              Database user
    DB_PASSWORD          Database password
    DB_NAME              Database name
EOF
}

# Get database connection string
get_db_url() {
    if [ -n "$DATABASE_URL" ]; then
        echo "$DATABASE_URL"
        return
    fi

    local host="${DB_HOST:-localhost}"
    local port="${DB_PORT:-5432}"
    local user="${DB_USER:-postgres}"
    local pass="${DB_PASSWORD}"
    local name="${DB_NAME:-functionfly}"
    local sslmode="${DB_SSLMODE:-disable}"

    if [ -z "$pass" ]; then
        read -s -p "Database password: " pass
        echo ""
    fi

    echo "postgres://${user}:${pass}@${host}:${port}/${name}?sslmode=${sslmode}"
}

# Check if migration table exists
check_migration_table() {
    local db_url="$1"

    psql "$db_url" -t -c "SELECT 1 FROM pg_tables WHERE tablename = 'schema_migrations'" 2>/dev/null | grep -q 1
}

# Get current migration version
get_current_version() {
    local db_url="$1"

    if ! check_migration_table "$db_url" "$2"; then
        echo "0"
        return
    fi

    psql "$db_url" -t -c "SELECT COALESCE(MAX(version), 0) FROM schema_migrations" 2>/dev/null | tr -d ' '
}

# Get all applied migrations
get_applied_migrations() {
    local db_url="$1"

    if ! check_migration_table "$db_url"; then
        echo "()"
        return
    fi

    psql "$db_url" -t -c "SELECT version FROM schema_migrations ORDER BY applied_at DESC" 2>/dev/null | tr -d ' '
}

# Get pending migrations
get_pending_migrations() {
    local current_version="$1"

    local pending=()
    for f in "${MIGRATIONS_DIR}"/*.up.sql; do
        if [ -f "$f" ]; then
            local version
            version=$(basename "$f" .up.sql)
            if [ "$version" -gt "$current_version" ]; then
                pending+=("$version")
            fi
        fi
    done

    echo "${pending[*]}"
}

# Roll back one migration
rollback_one() {
    local db_url="$1"
    local dry_run="$2"

    local current_version
    current_version=$(get_current_version "$db_url")

    if [ "$current_version" = "0" ] || [ -z "$current_version" ]; then
        log_warn "No migrations to roll back (schema_migrations table empty or missing)"
        return 0
    fi

    # Find the migration file
    local migration_file="${MIGRATIONS_DIR}/${current_version}_*.down.sql"

    # Try exact match first
    if [ -f "${MIGRATIONS_DIR}/${current_version}.down.sql" ]; then
        migration_file="${MIGRATIONS_DIR}/${current_version}.down.sql"
    else
        # Find by pattern
        local matched
        matched=$(ls "${MIGRATIONS_DIR}"/${current_version}_*.down.sql 2>/dev/null | head -1)
        if [ -n "$matched" ]; then
            migration_file="$matched"
        fi
    fi

    if [ ! -f "$migration_file" ]; then
        log_error "No down migration found for version ${current_version}"
        log_error "Expected file: ${MIGRATIONS_DIR}/${current_version}.down.sql or pattern"
        return 1
    fi

    log_info "Rolling back migration: ${current_version}"
    if [ "$dry_run" = "1" ]; then
        log_info "DRY RUN - SQL that would be executed:"
        echo ""
        cat "$migration_file"
        echo ""
        return 0
    fi

    log_info "Executing rollback migration..."

    # Execute the down migration
    if psql "$db_url" -f "$migration_file" 2>&1; then
        # Update schema_migrations table
        psql "$db_url" -c "DELETE FROM schema_migrations WHERE version = ${current_version}" 2>&1

        log_info "Successfully rolled back migration ${current_version}"
        return 0
    else
        log_error "Failed to roll back migration ${current_version}"
        return 1
    fi
}

# Roll back to specific version
rollback_to_version() {
    local db_url="$1"
    local target_version="$2"
    local dry_run="$3"

    local current_version
    current_version=$(get_current_version "$db_url")

    if [ "$current_version" -le "$target_version" ]; then
        log_warn "Current version ${current_version} is not greater than target ${target_version}"
        log_warn "Nothing to roll back"
        return 0
    fi

    log_info "Rolling back from version ${current_version} to ${target_version}..."

    while [ "$current_version" -gt "$target_version" ]; do
        if ! rollback_one "$db_url" "$dry_run"; then
            log_error "Rollback failed at version ${current_version}"
            return 1
        fi
        current_version=$(get_current_version "$db_url")
    done

    log_info "Successfully rolled back to version ${target_version}"
    return 0
}

# Show migration status
show_status() {
    local db_url="$1"

    echo ""
    echo "=========================================="
    echo "       Migration Status"
    echo "=========================================="
    echo ""

    # Check if migration table exists
    if ! check_migration_table "$db_url"; then
        log_warn "schema_migrations table does not exist"
        log_info "Run migrations to initialize the database schema"
        return 0
    fi

    local current_version
    current_version=$(get_current_version "$db_url")

    echo "Current version: ${current_version}"
    echo ""

    # Show applied migrations
    echo "Applied migrations:"
    local applied
    applied=$(get_applied_migrations "$db_url")
    if [ -n "$applied" ] && [ "$applied" != "()" ]; then
        echo "$applied" | tr ' ' '\n' | while read -r ver; do
            echo "  ✓ $ver"
        done
    else
        echo "  (none)"
    fi
    echo ""

    # Show pending migrations
    echo "Pending migrations:"
    local pending
    pending=$(get_pending_migrations "$current_version")
    if [ -n "$pending" ]; then
        echo "$pending" | tr ' ' '\n' | while read -r ver; do
            echo "  ○ $ver"
        done
    else
        echo "  (none - all up to date)"
    fi
    echo ""
}

# List all migrations
list_migrations() {
    echo ""
    echo "=========================================="
    echo "       Available Migrations"
    echo "=========================================="
    echo ""

    local current_version
    current_version=$(get_current_version "$(get_db_url)")

    echo "Migration files in ${MIGRATIONS_DIR}:"
    echo ""
    for f in "${MIGRATIONS_DIR}"/*.up.sql; do
        if [ -f "$f" ]; then
            local version
            version=$(basename "$f" .up.sql)
            local status="pending"
            if [ "$version" -le "$current_version" ]; then
                status="applied"
            fi
            printf "  [%s] %s (%s)\n" "$version" "$(basename "$f")" "$status"
        fi
    done | sort -t'[' -k2 -n
    echo ""
}

# Main
main() {
    if [ $# -eq 0 ]; then
        show_usage
        exit 1
    fi

    local db_url
    db_url=$(get_db_url)

    local command="$1"
    shift

    case "$command" in
        status)
            show_status "$db_url"
            ;;
        down)
            rollback_one "$db_url" "$DRY_RUN"
            ;;
        version)
            if [ $# -eq 0 ]; then
                log_error "Version argument required"
                show_usage
                exit 1
            fi
            rollback_to_version "$db_url" "$1" "$DRY_RUN"
            ;;
        list)
            list_migrations
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            log_error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
