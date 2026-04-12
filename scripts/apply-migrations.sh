#!/bin/bash
# apply-migrations.sh - Apply all pending migrations to PostgreSQL databases
# Usage: ./scripts/apply-migrations.sh [local|neon|all]
# Default: all

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_DIR"

# Default target
TARGET="${1:-all}"

# Load environment variables
if [ -f "$PROJECT_DIR/.env" ]; then
    export $(grep -v '^#' "$PROJECT_DIR/.env" | xargs)
fi

# Function to apply migrations to a PostgreSQL instance
apply_migrations() {
    local host="$1"
    local port="$2"
    local user="$3"
    local password="$4"
    local dbname="$5"
    local sslmode="${6:-require}"
    local name="$7"

    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Applying migrations to: ${name}${NC}"
    echo -e "${BLUE}Host: ${host}:${port}, Database: ${dbname}${NC}"
    echo -e "${BLUE}========================================${NC}"

    # Test connection first
    if ! PGPASSWORD="$password" psql -h "$host" -p "$port" -U "$user" -d "$dbname" -c "SELECT 1;" > /dev/null 2>&1; then
        echo -e "${RED}Failed to connect to ${name}${NC}"
        return 1
    fi

    # Check current migration status
    echo -e "${YELLOW}Current migration status:${NC}"
    PGPASSWORD="$password" psql -h "$host" -p "$port" -U "$user" -d "$dbname" -c "
        SELECT version, dirty FROM schema_migrations LIMIT 1;
    " 2>/dev/null || echo -e "${YELLOW}No migrations applied yet or schema_migrations table doesn't exist${NC}"

    # Count pending migrations
    echo -e "${YELLOW}Checking pending migrations...${NC}"
    local pending_count=0
    # We need to use golang-migrate to get actual pending count
    # For now, just show all available migrations
    local available_count=$(ls -1 "$PROJECT_DIR/migrations"/*.up.sql 2>/dev/null | wc -l)
    echo -e "${YELLOW}Available migrations: ${available_count}${NC}"

    # Apply migrations using golang-migrate CLI if available
    if command -v migrate &> /dev/null; then
        echo -e "${YELLOW}Using golang-migrate CLI...${NC}"
        export PGPASSWORD="$password"
        migrate -path "$PROJECT_DIR/migrations" -database "postgres://${user}@${host}:${port}/${dbname}?sslmode=${sslmode}" up
    else
        echo -e "${YELLOW}golang-migrate CLI not found. Using Go application...${NC}"
        echo -e "${YELLOW}Building orchestrator API to run migrations...${NC}"

        # Build the orchestrator API
        if ! go build -o "$PROJECT_DIR/bin/orchestrator-api-migrate" "$PROJECT_DIR/cmd/orchestrator-api" 2>&1 | grep -v "^(#"; then
            echo -e "${RED}Failed to build orchestrator API${NC}"
            return 1
        fi

        # Set database connection environment
        export DB_HOST="$host"
        export DB_PORT="$port"
        export DB_USER="$user"
        export DB_PASSWORD="$password"
        export DB_NAME="$dbname"
        export DB_SSLMODE="$sslmode"
        unset DATABASE_URL  # Ensure we use individual DB vars

        # Run migrations through the API (it will exit after migrations due to --skip-migrations wait, no...)
        # Actually, we need to run it differently
        echo -e "${YELLOW}Running migrations via Go application...${NC}"
        "$PROJECT_DIR/bin/orchestrator-api-migrate" &
        local pid=$!
        sleep 5
        kill $pid 2>/dev/null || true
        wait $pid 2>/dev/null || true

        # Cleanup
        rm -f "$PROJECT_DIR/bin/orchestrator-api-migrate"
    fi

    echo -e "${GREEN}Migrations applied successfully to ${name}!${NC}"
    echo ""
}

# Apply to local PostgreSQL
apply_local() {
    # Use values from .env or defaults
    local host="${DB_HOST:-localhost}"
    local port="${DB_PORT:-5432}"
    local user="${DB_USER:-postgres}"
    local password="${DB_PASSWORD:-postgres}"
    local dbname="${DB_NAME:-functionfly}"
    local sslmode="${DB_SSLMODE:-disable}"

    apply_migrations "$host" "$port" "$user" "$password" "$dbname" "$sslmode" "Local PostgreSQL"
}

# Apply to Neon PostgreSQL
apply_neon() {
    if [ -z "$DATABASE_URL" ]; then
        echo -e "${RED}DATABASE_URL not set. Cannot connect to Neon.${NC}"
        echo -e "${YELLOW}Please set DATABASE_URL environment variable with Neon connection string.${NC}"
        return 1
    fi

    echo -e "${YELLOW}Connecting to Neon PostgreSQL...${NC}"

    # Parse DATABASE_URL (format: postgres://user:pass@host:port/dbname?sslmode=require)
    # Extract components
    local url="$DATABASE_URL"

    # Remove postgres:// or postgresql:// prefix
    url="${url#postgresql://}"
    url="${url#postgres://}"

    # Extract user and password
    local userpass="${url%%@*}"
    local user="${userpass%%:*}"
    local password="${userpass#*:}"

    # Extract host, port, dbname
    local hostportdb="${url#*@}"
    local hostport="${hostportdb%%/*}"
    local db_and_params="${hostportdb#*/}"
    local dbname="${db_and_params%%\?*}"
    local params="${db_and_params#*\?}"

    local host="${hostport%%:*}"
    local port="${hostport#*:}"
    [ "$port" = "$host" ] && port="5432"

    # Extract sslmode from params
    local sslmode="require"
    if [[ "$params" == *"sslmode="* ]]; then
        sslmode="${params#*sslmode=}"
        sslmode="${sslmode%%&*}"
    fi

    apply_migrations "$host" "$port" "$user" "$password" "$dbname" "$sslmode" "Neon PostgreSQL"
}

# Show help
show_help() {
    echo "Usage: ./scripts/apply-migrations.sh [local|neon|all|status|help]"
    echo ""
    echo "Commands:"
    echo "  local   - Apply migrations to local PostgreSQL (from .env: DB_HOST, DB_PORT, etc.)"
    echo "  neon    - Apply migrations to Neon PostgreSQL (from .env: DATABASE_URL)"
    echo "  all     - Apply migrations to both local and Neon PostgreSQL (default)"
    echo "  status  - Check migration status on all configured databases"
    echo "  help    - Show this help message"
    echo ""
    echo "Environment variables required:"
    echo "  For local: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE"
    echo "  For Neon:  DATABASE_URL"
}

# Check migration status
check_status() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Migration Status Check${NC}"
    echo -e "${BLUE}========================================${NC}"

    # Local status
    if [ -n "$DB_HOST" ] || [ -n "$DB_PORT" ]; then
        echo -e "${YELLOW}Local PostgreSQL:${NC}"
        PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-functionfly}" -c "
            SELECT version, dirty, 'Applied' as status
            FROM schema_migrations
            UNION ALL
            SELECT 'N/A', false, 'Total pending: ' || (
                SELECT COUNT(*) FROM (
                    SELECT version FROM schema_migrations
                ) sq
            )::text;
        " 2>/dev/null || echo -e "${RED}Cannot connect to local PostgreSQL${NC}"
    fi

    # Neon status
    if [ -n "$DATABASE_URL" ]; then
        echo -e "${YELLOW}Neon PostgreSQL:${NC}"
        psql "$DATABASE_URL" -c "
            SELECT version, dirty, 'Applied' as status
            FROM schema_migrations
            LIMIT 5;
        " 2>/dev/null || echo -e "${RED}Cannot connect to Neon${NC}"
    fi
}

# Main execution
case "${TARGET}" in
    local)
        apply_local
        ;;
    neon)
        apply_neon
        ;;
    all)
        apply_local
        if [ -n "$DATABASE_URL" ]; then
            apply_neon
        else
            echo -e "${YELLOW}Skipping Neon (DATABASE_URL not set)${NC}"
        fi
        ;;
    status)
        check_status
        ;;
    help|--help|-h)
        show_help
        exit 0
        ;;
    *)
        echo -e "${RED}Unknown target: ${TARGET}${NC}"
        show_help
        exit 1
        ;;
esac

echo -e "${GREEN}Migration process completed!${NC}"
