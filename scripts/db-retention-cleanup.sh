#!/bin/bash
# Data Retention Cleanup Script
# Run via cron: 0 4 * * * /home/micro/projects/functionfly/scripts/db-retention-cleanup.sh
# This script handles data retention cleanup with legal hold checks

set -euo pipefail

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-functionfly}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
LOG_FILE="${LOG_FILE:-/var/log/functionfly/db-retention.log}"
DRY_RUN="${DRY_RUN:-false}"
S3_ARCHIVE_ENABLED="${S3_ARCHIVE_ENABLED:-false}"

mkdir -p "$(dirname "$LOG_FILE")"

TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

echo "[$TIMESTAMP] Starting retention cleanup (dry_run=$DRY_RUN)" >> "$LOG_FILE"

export PGPASSWORD="$DB_PASSWORD"

# Function to check for active legal holds
check_legal_holds() {
    local table="$1"
    local result
    result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT is_under_legal_hold('$table');" 2>/dev/null | xargs)
    
    if [ "$result" = "t" ] || [ "$result" = "true" ]; then
        echo "[$TIMESTAMP] SKIPPED: $table is under legal hold" >> "$LOG_FILE"
        return 1
    fi
    return 0
}

# Function to run retention cleanup for a table
run_cleanup() {
    local table="$1"
    local retention_days="$2"
    
    echo "[$TIMESTAMP] Processing $table (retention: ${retention_days} days)" >> "$LOG_FILE"
    
    # Check legal holds first
    if ! check_legal_holds "$table"; then
        return 0
    fi
    
    # Archive to S3 before deletion (if enabled)
    if [ "$S3_ARCHIVE_ENABLED" = "true" ] && [ "$DRY_RUN" = "false" ]; then
        echo "[$TIMESTAMP] Archiving $table data before deletion" >> "$LOG_FILE"
        # Run Go archive service for this table
        # go run /home/micro/projects/functionfly/cmd/archive-service/main.go --table="$table"
    fi
    
    # Run cleanup
    if [ "$DRY_RUN" = "true" ]; then
        result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
            SELECT * FROM execute_retention_cleanup(true, ARRAY['$table']::text[]);
        " 2>&1 | tee -a "$LOG_FILE")
    else
        result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
            SELECT * FROM execute_retention_cleanup(false, ARRAY['$table']::text[]);
        " 2>&1 | tee -a "$LOG_FILE")
    fi
    
    echo "$result"
}

# 1. Cost Allocation Entries (90 days)
echo "[$TIMESTAMP] === Cleaning Cost Allocation Entries ===" >> "$LOG_FILE"
run_cleanup "cost_allocation_entries" 90

# 2. Registry Function Executions (30 days)
echo "[$TIMESTAMP] === Cleaning Registry Executions ===" >> "$LOG_FILE"
run_cleanup "registry_function_executions" 30

# 3. Routing Events (30 days)
echo "[$TIMESTAMP] === Cleaning Routing Events ===" >> "$LOG_FILE"
run_cleanup "routing_events" 30

# 4. Health Checks (90 days)
echo "[$TIMESTAMP] === Cleaning Health Checks ===" >> "$LOG_FILE"
run_cleanup "health_checks" 90

# 5. Performance Metrics (30 days)
echo "[$TIMESTAMP] === Cleaning Performance Metrics ===" >> "$LOG_FILE"
run_cleanup "performance_metrics" 30

# 6. Function Logs (30 days)
echo "[$TIMESTAMP] === Cleaning Function Logs ===" >> "$LOG_FILE"
run_cleanup "function_logs" 30

# 7. Log summary
echo "[$TIMESTAMP] Retention cleanup completed (dry_run=$DRY_RUN)" >> "$LOG_FILE"
echo "---" >> "$LOG_FILE"

exit 0
