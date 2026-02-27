#!/bin/bash

# PostgreSQL load testing script using pgbench
# This script provides comprehensive database load testing capabilities

set -euo pipefail

ENVIRONMENT=${1:-production}
SCALE_FACTOR=${2:-50}  # Database size multiplier (50 = ~750MB database)
CLIENTS=${3:-10}      # Number of concurrent clients
THREADS=${4:-2}       # Number of worker threads
DURATION=${5:-60}     # Test duration in seconds
TEST_TYPE=${6:-tpcb}  # Test type: tpcb, select, update, mixed

# Load environment variables
if [ -f ".env.${ENVIRONMENT}" ]; then
    source ".env.${ENVIRONMENT}"
elif [ -f "deploy/database/${ENVIRONMENT}.env" ]; then
    source "deploy/database/${ENVIRONMENT}.env"
fi

# Database connection details
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-functionfly}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}PostgreSQL Load Testing Script${NC}"
echo "Environment: $ENVIRONMENT"
echo "Scale Factor: $SCALE_FACTOR"
echo "Clients: $CLIENTS"
echo "Threads: $THREADS"
echo "Duration: ${DURATION}s"
echo "Test Type: $TEST_TYPE"
echo

# Function to run pgbench
run_pgbench() {
    local test_name=$1
    local extra_args=$2

    echo -e "${YELLOW}Running $test_name test...${NC}"

    PGPASSWORD="$DB_PASSWORD" pgbench \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --username="$DB_USER" \
        --dbname="$DB_NAME" \
        --scale="$SCALE_FACTOR" \
        --clients="$CLIENTS" \
        --threads="$THREADS" \
        --time="$DURATION" \
        --progress=10 \
        --report-per-command \
        $extra_args
}

# Initialize test database if needed
initialize_db() {
    echo -e "${BLUE}Initializing pgbench database...${NC}"

    # Drop and recreate test database
    PGPASSWORD="$DB_PASSWORD" psql \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --username="$DB_USER" \
        --dbname="postgres" \
        -c "DROP DATABASE IF EXISTS ${DB_NAME}_bench;" \
        -c "CREATE DATABASE ${DB_NAME}_bench;"

    # Initialize pgbench tables
    PGPASSWORD="$DB_PASSWORD" pgbench \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --username="$DB_USER" \
        --dbname="${DB_NAME}_bench" \
        --initialize \
        --scale="$SCALE_FACTOR"

    echo -e "${GREEN}Database initialized successfully${NC}"
}

# Run different types of tests
case "$TEST_TYPE" in
    "init")
        initialize_db
        ;;

    "tpcb")
        echo "Running TPC-B (Transaction Processing Performance Council) benchmark..."
        echo "This simulates a typical OLTP workload with:"
        echo "- Account balance updates"
        echo "- Teller transfers"
        echo "- Branch summary reports"
        echo
        run_pgbench "TPC-B" ""
        ;;

    "select")
        echo "Running SELECT-only benchmark..."
        run_pgbench "SELECT" "--select-only"
        ;;

    "update")
        echo "Running UPDATE-only benchmark..."
        run_pgbench "UPDATE" "--skip-some-updates"
        ;;

    "mixed")
        echo "Running mixed workload benchmark..."
        # Custom script for mixed operations
        cat > /tmp/mixed_benchmark.sql << 'EOF'
\set aid random(1, 100000 * :scale)
\set bid random(1, 1 * :scale)
\set tid random(1, 10 * :scale)
\set delta random(-5000, 5000)
BEGIN;
UPDATE pgbench_accounts SET abalance = abalance + :delta WHERE aid = :aid;
SELECT abalance FROM pgbench_accounts WHERE aid = :aid;
UPDATE pgbench_tellers SET tbalance = tbalance + :delta WHERE tid = :tid;
UPDATE pgbench_branches SET bbalance = bbalance + :delta WHERE bid = :bid;
INSERT INTO pgbench_history (tid, bid, aid, delta, mtime) VALUES (:tid, :bid, :aid, :delta, CURRENT_TIMESTAMP);
END;
EOF
        run_pgbench "Mixed" "--file=/tmp/mixed_benchmark.sql"
        rm -f /tmp/mixed_benchmark.sql
        ;;

    "custom")
        echo "Running custom benchmark with your application's typical queries..."

        # Create custom benchmark script based on your app's query patterns
        cat > /tmp/custom_benchmark.sql << 'EOF'
-- Custom benchmark simulating your application's query patterns
\set tenant_id random(1, 100)
\set user_id random(1, 1000)
\set app_id random(1, 100)

-- Simulate user authentication (common query)
SELECT id, email, password_hash FROM users WHERE email = 'user' || :user_id || '@example.com';

-- Simulate app listing (common query)
SELECT id, name, slug FROM apps WHERE tenant_id = :tenant_id LIMIT 20;

-- Simulate deployment queries (common query)
SELECT id, status, created_at FROM deployments WHERE app_id = :app_id ORDER BY created_at DESC LIMIT 10;

-- Simulate audit logging (write-heavy operation)
INSERT INTO audit_events (actor_email, tenant_id, action, resource_type, timestamp, success)
VALUES ('user' || :user_id || '@example.com', :tenant_id, 'api_call', 'function', NOW(), true);

-- Simulate usage tracking (write-heavy operation)
INSERT INTO usage_events (tenant_id, event_type, quantity, timestamp)
VALUES (:tenant_id, 'api_call', 1, NOW());
EOF

        # Initialize the benchmark database with your schema first
        echo "Setting up custom benchmark database..."
        PGPASSWORD="$DB_PASSWORD" psql \
            --host="$DB_HOST" \
            --port="$DB_PORT" \
            --username="$DB_USER" \
            --dbname="${DB_NAME}_bench" \
            -f "internal/storage/sql/migrations/20240101000000_initial_schema.up.sql" || true

        run_pgbench "Custom Application" "--file=/tmp/custom_benchmark.sql"
        rm -f /tmp/custom_benchmark.sql
        ;;

    *)
        echo -e "${RED}Invalid test type: $TEST_TYPE${NC}"
        echo "Available test types:"
        echo "  init   - Initialize benchmark database"
        echo "  tpcb   - Standard TPC-B benchmark"
        echo "  select - SELECT-only benchmark"
        echo "  update - UPDATE-only benchmark"
        echo "  mixed  - Mixed read/write benchmark"
        echo "  custom - Custom benchmark based on your app's queries"
        exit 1
        ;;
esac

echo
echo -e "${GREEN}Load testing completed!${NC}"
echo
echo "Performance Analysis Tips:"
echo "1. TPS (Transactions Per Second): Higher is better"
echo "2. Latency: Lower is better (aim for < 10ms)"
echo "3. Connection pooling: Ensure you're not exhausting connections"
echo "4. Monitor system resources: CPU, memory, disk I/O"
echo "5. Compare results with different client/thread counts"
echo
echo "For continuous monitoring during tests, run in another terminal:"
echo "  watch -n 1 'psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c \"SELECT * FROM pg_stat_activity;\"'"