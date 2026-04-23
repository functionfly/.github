#!/bin/bash
# Database Maintenance Script for FunctionFly
# This script performs routine maintenance to keep database performance optimal
# Run via cron: 0 3 * * * /home/micro/projects/functionfly/scripts/db-maintenance.sh

set -euo pipefail

# Database connection parameters
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-functionfly}
LOG_FILE=${LOG_FILE:-/var/log/functionfly/db-maintenance.log}
SLACK_WEBHOOK=${SLACK_WEBHOOK:-}

# Special handling for Docker Postgres (common in development)
if [ "$DB_PORT" = "5432" ] && ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
    # Try Docker port if default port fails
    if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "5434" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
        echo "Using Docker Postgres on port 5434"
        DB_PORT=5434
    fi
fi

# Create log directory if needed
mkdir -p "$(dirname "$LOG_FILE")" 2>/dev/null || true

TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$TIMESTAMP] 🛠️  Starting database maintenance..."
echo "[$TIMESTAMP]    Host: $DB_HOST:$DB_PORT"
echo "[$TIMESTAMP]    Database: $DB_NAME"
echo "[$TIMESTAMP]    User: $DB_USER"

# Export password for psql
export PGPASSWORD="$DB_PASSWORD"

# Function to run maintenance commands with logging
run_maintenance() {
    local command="$1"
    local description="$2"

    echo "[$TIMESTAMP] 📋 $description..."
    
    if result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$command" -q 2>&1); then
        echo "[$TIMESTAMP] ✅ $description completed"
        [ -n "$result" ] && echo "[$TIMESTAMP] Result: $result"
    else
        echo "[$TIMESTAMP] ⚠️  $description failed: $result"
    fi
}

# Function to log to file
log_to_file() {
    echo "$1" >> "$LOG_FILE" 2>/dev/null || true
}

log_to_file "[$TIMESTAMP] Starting database maintenance"

# 1. Check database health first
echo "[$TIMESTAMP] 🔍 Checking database health..."
health_result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT jsonb_agg(jsonb_build_object('check', check_name, 'status', status)) FROM quick_health_check();" 2>/dev/null || echo "[]")

# Check for critical issues
critical_count=$(echo "$health_result" | grep -c "CRITICAL" || echo "0")
warning_count=$(echo "$health_result" | grep -c "WARNING" || echo "0")

if [ "$critical_count" -gt 0 ]; then
    echo "[$TIMESTAMP] 🚨 $critical_count CRITICAL issues found!"
    log_to_file "[$TIMESTAMP] CRITICAL: $critical_count critical issues"
    if [ -n "$SLACK_WEBHOOK" ]; then
        curl -s -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"🚨 Database CRITICAL: $critical_count issues on $HOSTNAME - check logs\"}" \
            "$SLACK_WEBHOOK" 2>/dev/null || true
    fi
fi

if [ "$warning_count" -gt 0 ]; then
    echo "[$TIMESTAMP] ⚠️  $warning_count warnings found"
fi

# 2. Refresh materialized views (if function exists)
run_maintenance "SELECT refresh_billing_materialized_views();" "Refreshing billing materialized views"

# 3. Partition maintenance (if functions exist)
run_maintenance "SELECT create_next_month_partitions_native();" "Creating next month partitions"

# 4. Drop old partitions
echo "[$TIMESTAMP] 🗑️  Dropping old partitions..."
run_maintenance "SELECT drop_old_partitions_native('function_logs', 30);" "Drop old function log partitions"
run_maintenance "SELECT drop_old_partitions_native('performance_metrics', 30);" "Drop old performance metrics"
run_maintenance "SELECT drop_old_partitions_native('routing_events', 30);" "Drop old routing events"

# 5. Vacuum bloated tables (dry run first, then real)
echo "[$TIMESTAMP] 🧹 Vacuum maintenance..."
run_maintenance "SELECT vacuum_bloated_tables(10000, true);" "Preview bloated tables (dry run)"

# 6. Analyze key tables for better query planning
run_maintenance "ANALYZE registry_functions, registry_function_ratings, registry_function_verification_status, registry_function_versions, cost_allocation_entries, registry_function_executions;" "Analyzing table statistics"

# 7. Vacuum analyze to clean up bloat and update statistics
run_maintenance "VACUUM ANALYZE registry_functions;" "Vacuuming registry_functions table"
run_maintenance "VACUUM ANALYZE registry_function_ratings;" "Vacuuming registry_function_ratings table"
run_maintenance "VACUUM ANALYZE cost_allocation_entries;" "Vacuuming cost_allocation_entries"

# 8. Check for unused indexes
echo "[$TIMESTAMP] 🔍 Checking for unused indexes..."
unused_indexes=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
SELECT schemaname || '.' || relname || '.' || indexrelname as full_name
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
AND idx_scan = 0
AND indexrelname NOT LIKE 'uq_%'
ORDER BY pg_relation_size(indexrelid) DESC
LIMIT 10;
" 2>/dev/null || true)

if [ -n "$unused_indexes" ]; then
    echo "[$TIMESTAMP]    Unused indexes found:"
    echo "$unused_indexes" | while read -r line; do
        echo "[$TIMESTAMP]      - $line"
    done
fi

# 9. Check replication lag if replicas exist
replication_lag=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
SELECT count(*) FROM check_replication_lag() WHERE status IN ('WARNING', 'CRITICAL');
" 2>/dev/null || echo "0")

if [ "$replication_lag" -gt 0 ] 2>/dev/null; then
    echo "[$TIMESTAMP] ⚠️  $replication_lag replica(s) with lag issues"
fi

TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo "[$TIMESTAMP] 🎉 Database maintenance completed successfully!"
echo "[$TIMESTAMP] ---"

log_to_file "[$TIMESTAMP] Maintenance completed"
log_to_file "---"

echo ""
echo "💡 Tips to maintain performance:"
echo "   - Run this script weekly or after large data imports"
echo "   - Monitor slow queries with: SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
echo "   - Consider adding more indexes if specific queries remain slow"
echo "   - Check retention_summary view for cleanup status"
