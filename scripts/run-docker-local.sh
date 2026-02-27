#!/usr/bin/env bash
# Start the local Docker stack (everything except Postgres).
# Postgres must be running on the host (e.g. localhost:5432 or 5434 if using Docker Postgres). Migrations are not run automatically.
#
# Optional env: DB_HOST, DB_PORT (default 5432 for host Postgres), DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE
#
# Usage:
#   ./scripts/run-docker-local.sh [compose up options, e.g. -d]
set -e
cd "$(dirname "$0")/.."

export DB_HOST="${DB_HOST:-host.docker.internal}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-functionfly}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"

echo "Starting Docker stack (no Postgres). DB: $DB_HOST:$DB_PORT/$DB_NAME"
exec docker compose -f docker-compose.local.yml up "$@"
