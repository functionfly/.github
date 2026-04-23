#!/bin/bash
# Database Health Check Reporter
# Run via cron: */15 * * * * /home/micro/projects/functionfly/scripts/db-health-check.sh
# Reports database health metrics for monitoring

set -euo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-functionfly}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
METRICS_FILE="${METRICS_FILE:-/tmp/db-health-metrics.json}"
export PGPASSWORD="$DB_PASSWORD"

# Get health check data
health_data=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
    SELECT jsonb_pretty(jsonb_agg(jsonb_build_object(
        'check_name', check_name,
        'status', status,
        'details', details
    ))) FROM quick_health_check();
" 2>/dev/null | tr -d '\n')

# Get replication lag
replication_lag=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
    SELECT jsonb_pretty(jsonb_agg(jsonb_build_object(
        'replica', replica_address,
        'lag', lag_size,
        'status', status
    ))) FROM check_replication_lag();
" 2>/dev/null | tr -d '\n')

# Get slow queries
slow_queries=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
    SELECT jsonb_pretty(jsonb_agg(jsonb_build_object(
        'query', query_preview,
        'mean_time', mean_time_ms,
        'calls', calls
    ))) FROM slow_queries LIMIT 5;
" 2>/dev/null | tr -d '\n')

# Get connection count
conn_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
    SELECT count(*) FROM pg_stat_activity;
" 2>/dev/null | tr -d '\n')

# Build metrics JSON
cat > "$METRICS_FILE" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "database": "$DB_NAME",
    "host": "$DB_HOST",
    "connections": $conn_count,
    "health_checks": $health_data,
    "replication_lag": $replication_lag,
    "slow_queries": $slow_queries
}
EOF

# Output to stdout for log collection
cat "$METRICS_FILE"

exit 0
