#!/bin/bash

# Point-in-time recovery script for FunctionFly PostgreSQL
# Usage: ./restore-database-pitr.sh [environment] [target_time] [backup_file]

set -euo pipefail

ENVIRONMENT=${1:-production}
TARGET_TIME=${2:-}  # ISO timestamp for point-in-time recovery
BACKUP_FILE=${3:-}  # Base backup file to restore from

if [ -z "$TARGET_TIME" ] || [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 [environment] [target_time] [backup_file]"
    echo "Example: $0 production '2024-02-20 15:30:00+00' /path/to/basebackup.tar.gz"
    exit 1
fi

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

# Recovery configuration
RECOVERY_DIR="/tmp/pitr_recovery_$(date +%s)"
WAL_ARCHIVE_DIR="/var/backups/functionfly/wal_archive"

echo "Starting point-in-time recovery to: $TARGET_TIME"
echo "Using base backup: $BACKUP_FILE"
echo "Recovery directory: $RECOVERY_DIR"

# Create recovery directory
mkdir -p "$RECOVERY_DIR"

# Stop the database (in production, you'd coordinate this carefully)
echo "Stopping database for recovery..."
docker-compose -f docker-compose.production.yml down postgres

# Extract base backup
echo "Extracting base backup..."
tar -xzf "$BACKUP_FILE" -C "$RECOVERY_DIR"

# Create recovery configuration
cat > "$RECOVERY_DIR/recovery.conf" << EOF
# Recovery configuration for point-in-time recovery
restore_command = 'cp $WAL_ARCHIVE_DIR/%f %p'
recovery_target_time = '$TARGET_TIME'
recovery_target_action = 'promote'
recovery_target_inclusive = false

# Logging
log_recovery_conflict_waits = on
EOF

# Create postgresql.conf for recovery
cat > "$RECOVERY_DIR/postgresql.conf" << EOF
# Recovery-specific configuration
listen_addresses = '*'
port = $DB_PORT
max_connections = 100

# Disable WAL archiving during recovery
wal_level = replica
archive_mode = off

# Memory settings
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# Logging
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_statement = 'ddl'
log_duration = on
log_min_duration_statement = 1000
log_checkpoints = on
EOF

# Copy WAL files to pg_wal directory if they exist
if [ -d "$WAL_ARCHIVE_DIR" ]; then
    echo "Copying archived WAL files..."
    mkdir -p "$RECOVERY_DIR/pg_wal"
    find "$WAL_ARCHIVE_DIR" -name "*.gz" -exec gunzip -c {} \; > "$RECOVERY_DIR/pg_wal/backup_wal.tar"
    cd "$RECOVERY_DIR/pg_wal" && tar -xf backup_wal.tar && rm backup_wal.tar
fi

# Start PostgreSQL in recovery mode
echo "Starting PostgreSQL in recovery mode..."
export PGDATA="$RECOVERY_DIR"
export PGPASSWORD="$DB_PASSWORD"

# Initialize recovery
/usr/lib/postgresql/15/bin/pg_ctl -D "$RECOVERY_DIR" -l "$RECOVERY_DIR/recovery.log" start

# Wait for recovery to complete
echo "Waiting for recovery to complete..."
TIMEOUT=300
COUNTER=0
while [ $COUNTER -lt $TIMEOUT ]; do
    if /usr/lib/postgresql/15/bin/pg_isready -h localhost -p $DB_PORT >/dev/null 2>&1; then
        # Check if recovery is complete
        RECOVERY_STATUS=$(/usr/lib/postgresql/15/bin/psql -h localhost -p $DB_PORT -U $DB_USER -d postgres -t -c "SELECT pg_is_in_recovery();" 2>/dev/null || echo "error")
        if [ "$RECOVERY_STATUS" = "f" ]; then
            echo "Recovery completed successfully!"
            break
        fi
    fi

    sleep 5
    COUNTER=$((COUNTER + 5))
    echo "Recovery in progress... ($COUNTER/$TIMEOUT seconds)"
done

if [ $COUNTER -ge $TIMEOUT ]; then
    echo "Recovery timed out"
    /usr/lib/postgresql/15/bin/pg_ctl -D "$RECOVERY_DIR" stop -m immediate
    exit 1
fi

# Stop PostgreSQL
/usr/lib/postgresql/15/bin/pg_ctl -D "$RECOVERY_DIR" stop

# Backup original data directory
ORIGINAL_DATA_DIR="/var/lib/postgresql/data"
BACKUP_DATA_DIR="/var/lib/postgresql/data_backup_$(date +%Y%m%d_%H%M%S)"
echo "Backing up original data directory to: $BACKUP_DATA_DIR"
mv "$ORIGINAL_DATA_DIR" "$BACKUP_DATA_DIR"

# Move recovered data to production location
echo "Moving recovered data to production location..."
mv "$RECOVERY_DIR" "$ORIGINAL_DATA_DIR"

# Fix permissions
chown -R postgres:postgres "$ORIGINAL_DATA_DIR"

# Start the database
echo "Starting recovered database..."
docker-compose -f docker-compose.production.yml up -d postgres

# Wait for database to be ready
echo "Waiting for database to be ready..."
TIMEOUT=60
COUNTER=0
while [ $COUNTER -lt $TIMEOUT ]; do
    if docker-compose -f docker-compose.production.yml exec -T postgres pg_isready -U $DB_USER >/dev/null 2>&1; then
        echo "Database is ready!"
        break
    fi
    sleep 2
    COUNTER=$((COUNTER + 2))
done

if [ $COUNTER -ge $TIMEOUT ]; then
    echo "Database failed to start after recovery"
    exit 1
fi

# Verify recovery point
echo "Verifying recovery point..."
RECOVERY_TIME=$(docker-compose -f docker-compose.production.yml exec -T postgres psql -U $DB_USER -d $DB_NAME -t -c "
    SELECT 'Recovery completed at: ' || now() || ' (target was: $TARGET_TIME)';")

echo "$RECOVERY_TIME"

# Run post-recovery validation
echo "Running post-recovery validation..."
docker-compose -f docker-compose.production.yml exec -T postgres psql -U $DB_USER -d $DB_NAME -c "
    SELECT 'Database size:' || pg_size_pretty(pg_database_size('$DB_NAME'));
    SELECT 'Table count:' || count(*) FROM information_schema.tables WHERE table_schema = 'public';
    SELECT 'User count:' || count(*) FROM users;
    SELECT 'Last transaction:' || max(timestamp) FROM audit_events;
"

echo ""
echo "Point-in-time recovery completed successfully!"
echo "Original data backed up to: $BACKUP_DATA_DIR"
echo "Recovered to target time: $TARGET_TIME"
echo ""
echo "To rollback if needed:"
echo "docker-compose -f docker-compose.production.yml down postgres"
echo "mv $ORIGINAL_DATA_DIR ${ORIGINAL_DATA_DIR}_failed_recovery"
echo "mv $BACKUP_DATA_DIR $ORIGINAL_DATA_DIR"
echo "docker-compose -f docker-compose.production.yml up -d postgres"