#!/usr/bin/env bash
# Reset the functionfly database: drop, create, run migrations.
# Use when the DB is in a broken or inconsistent state (e.g. dirty migrations, missing tables).
# Requires: PostgreSQL client (psql), DB_* env or defaults (see below).
set -e

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-functionfly}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

export PGPASSWORD="$DB_PASSWORD"

echo "Resetting database: $DB_NAME at $DB_HOST:$DB_PORT"
# Terminate existing connections so we can drop the database
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();
" 2>/dev/null || true
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS $DB_NAME;"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE $DB_NAME;"
echo "Running migrations..."
DB_HOST="$DB_HOST" DB_PORT="$DB_PORT" DB_USER="$DB_USER" DB_PASSWORD="$DB_PASSWORD" DB_NAME="$DB_NAME" DB_SSLMODE="$DB_SSLMODE" \
  go run ./cmd/migrate up
echo "Done. Run 'make setup' to create initial tenant and admin user if needed."
