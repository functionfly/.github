#!/bin/bash
# Database Maintenance Script for FunctionFly
# This script performs routine maintenance to keep database performance optimal

set -e

# Database connection parameters
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-functionfly}

# Special handling for Docker Postgres (common in development)
if [ "$DB_PORT" = "5432" ] && ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
    # Try Docker port if default port fails
    if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "5434" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
        echo "Using Docker Postgres on port 5434"
        DB_PORT=5434
    fi
fi

echo "🛠️  Starting database maintenance..."
echo "   Host: $DB_HOST:$DB_PORT"
echo "   Database: $DB_NAME"
echo "   User: $DB_USER"

# Export password for psql
export PGPASSWORD="$DB_PASSWORD"

# Function to run maintenance commands
run_maintenance() {
    local command="$1"
    local description="$2"

    echo "📋 $description..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$command" -q
    echo "✅ $description completed"
}

# Analyze key tables for better query planning
run_maintenance "ANALYZE registry_functions, registry_function_ratings, registry_function_verification_status, registry_function_versions;" "Analyzing table statistics"

# Vacuum analyze to clean up bloat and update statistics
run_maintenance "VACUUM ANALYZE registry_functions;" "Vacuuming registry_functions table"
run_maintenance "VACUUM ANALYZE registry_function_ratings;" "Vacuuming registry_function_ratings table"
run_maintenance "VACUUM ANALYZE registry_function_verification_status;" "Vacuuming registry_function_verification_status table"

# Check for unused indexes (optional, for manual review)
echo "🔍 Checking for unused indexes..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT schemaname, relname as table_name, indexrelname as index_name, idx_scan, pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
AND idx_scan = 0
AND indexrelname NOT LIKE 'uq_%'
ORDER BY pg_relation_size(indexrelid) DESC;
" -q || true

echo "🎉 Database maintenance completed successfully!"
echo ""
echo "💡 Tips to maintain performance:"
echo "   - Run this script weekly or after large data imports"
echo "   - Monitor slow queries with: SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"
echo "   - Consider adding more indexes if specific queries remain slow"
