#!/usr/bin/env bash
# setup-fwos-db.sh — Create and initialize the FWOS database
set -euo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
FWOS_DB_NAME="${FWOS_DB_NAME:-fwos}"

export PGPASSWORD="$DB_PASSWORD"

echo "Creating FWOS database..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -tc "SELECT 1 FROM pg_database WHERE datname = '$FWOS_DB_NAME'" | grep -q 1 || \
  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE $FWOS_DB_NAME"

echo "Running FWOS migrations..."
for migration in migrations/20260623*_fwos_*.up.sql migrations/20260623*_fwos_phase2*.up.sql; do
  if [ -f "$migration" ]; then
    echo "  Applying: $(basename "$migration")"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$FWOS_DB_NAME" -f "$migration"
  fi
done

echo "FWOS database ready: $FWOS_DB_NAME"
echo "Set FWOS_DB_NAME=$FWOS_DB_NAME or FWOS_DATABASE_URL=postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$FWOS_DB_NAME"
