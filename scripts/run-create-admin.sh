#!/usr/bin/env bash
# Run the create-admin Go command with DB env. Uses .env or dev defaults.
#
# Usage:
#   ./scripts/run-create-admin.sh
#   ADMIN_EMAIL=you@example.com ADMIN_PASSWORD=secret ./scripts/run-create-admin.sh
#   ./scripts/run-create-admin.sh -email you@example.com -password secret -role super_admin

set -e
cd "$(dirname "$0")/.."

# Load DB_* and ADMIN_* from .env if present; DB_PORT is never read from .env so script default wins unless caller sets it
if [ -f .env ]; then
  while IFS= read -r line; do
    key="${line%%=*}"
    case "$key" in
      DB_HOST|DB_USER|DB_PASSWORD|DB_NAME|DB_SSLMODE|ADMIN_EMAIL|ADMIN_PASSWORD|ADMIN_ROLE)
        [ -z "${!key}" ] && export "$line"
        ;;
      DB_PORT) ;; # skip: use script default 5432 or caller's DB_PORT
    esac
  done < <(grep -E '^[A-Za-z_][A-Za-z0-9_]*=' .env 2>/dev/null || true)
fi

# DB defaults: 5432 for local Postgres; set DB_PORT=5434 for Docker Postgres
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-functionfly}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

# SECURITY: Require ADMIN_CREATE_PASSWORD in production, use dev default only in DEVELOPMENT mode
if [ "$DEVELOPMENT" != "true" ]; then
    if [ -z "$ADMIN_PASSWORD" ] && [ -z "$ADMIN_CREATE_PASSWORD" ]; then
        echo "ERROR: In production, you must set ADMIN_PASSWORD or ADMIN_CREATE_PASSWORD env var."
        echo "       Alternatively, set DEVELOPMENT=true to use default dev credentials (NOT recommended for production)."
        exit 1
    fi
fi
EMAIL="${ADMIN_EMAIL:-admin@functionfly.com}"
PASSWORD="${ADMIN_PASSWORD:-${ADMIN_CREATE_PASSWORD:-}}"
ROLE="${ADMIN_ROLE:-super_admin}"

if [ $# -ge 1 ]; then
  exec go run ./cmd/create-admin/ "$@"
fi

echo "Creating admin user (DB: $DB_HOST:$DB_PORT/$DB_NAME)..."
echo "  Email: $EMAIL  Role: $ROLE"
go run ./cmd/create-admin/ -email "$EMAIL" -password "$PASSWORD" -role "$ROLE"
