#!/usr/bin/env bash
# Run the orchestrator API against local Postgres (no Infisical).
# Requires: Postgres on localhost (port 5432 for local, 5434 if using Docker postgres), DB migrated (or SKIP_MIGRATIONS=false).
set -e
cd "$(dirname "$0")/.."

export DB_HOST="${DB_HOST:-localhost}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-functionfly}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"
export PORT="${PORT:-8080}"
export JWT_SECRET="${JWT_SECRET}"
if [ -z "$JWT_SECRET" ]; then
    echo "ERROR: JWT_SECRET environment variable is required. Source your .env file or set JWT_SECRET explicitly."
    exit 1
fi
# Skip migrations when DB is already up-to-date so the server starts quickly
export SKIP_MIGRATIONS="${SKIP_MIGRATIONS:-true}"
export REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"
export DEVELOPMENT="${DEVELOPMENT:-true}"
# Optional: for Admin Registry "Generate with AI" description (Open Router)
export OPENROUTER_API_KEY="${OPENROUTER_API_KEY:-}"

echo "Starting API on port $PORT (DB: $DB_HOST:$DB_PORT/$DB_NAME, Redis: $REDIS_ADDR). SKIP_MIGRATIONS=$SKIP_MIGRATIONS"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
exec "$SCRIPT_DIR/run-orchestrator-with-ai.sh" go run ./cmd/orchestrator-api
