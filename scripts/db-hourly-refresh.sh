#!/bin/bash
# Hourly Analytics Refresh Script
# Run via cron: 0 * * * * /home/micro/projects/functionfly/scripts/db-hourly-refresh.sh

set -euo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-functionfly}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

export PGPASSWORD="$DB_PASSWORD"

# Refresh hourly materialized views
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT refresh_hourly_materialized_views();" 2>/dev/null || true

# Refresh regional metrics (4x daily - only on specific hours)
hour=$(date +%H)
if [ "$((hour % 6))" -eq 0 ]; then
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT refresh_regional_materialized_views();" 2>/dev/null || true
fi

exit 0
