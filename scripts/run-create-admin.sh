#!/usr/bin/env bash
# Run the create-admin Go command with DB env. Uses .env or dev defaults.
#
# Usage:
#   ./scripts/run-create-admin.sh
#   ADMIN_EMAIL=you@example.com ADMIN_PASSWORD=secret ./scripts/run-create-admin.sh
#   ./scripts/run-create-admin.sh -email you@example.com -password secret -role super_admin

set -e
cd "$(dirname "$0")/.."

# Load DB_* and ADMIN_* from .env if present (avoid sourcing whole file due to special chars)
if [ -f .env ]; then
  while IFS= read -r line; do
    case "$line" in
      DB_HOST=*|DB_PORT=*|DB_USER=*|DB_PASSWORD=*|DB_NAME=*|DB_SSLMODE=*|ADMIN_EMAIL=*|ADMIN_PASSWORD=*|ADMIN_ROLE=*)
        export "$line"
        ;;
    esac
  done < <(grep -E '^[A-Za-z_][A-Za-z0-9_]*=' .env 2>/dev/null || true)
fi

# DB defaults (match dev.sh / storage config)
export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5434}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-functionfly}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

# Default admin credentials; override with ADMIN_* env or pass -email/-password/-role to this script
EMAIL="${ADMIN_EMAIL:-admin@functionfly.local}"
PASSWORD="${ADMIN_PASSWORD:-admin123}"
ROLE="${ADMIN_ROLE:-super_admin}"

if [ $# -ge 1 ]; then
  exec go run ./cmd/create-admin/ "$@"
fi

echo "Creating admin user (DB: $DB_HOST:$DB_PORT/$DB_NAME)..."
echo "  Email: $EMAIL  Role: $ROLE"
go run ./cmd/create-admin/ -email "$EMAIL" -password "$PASSWORD" -role "$ROLE"
