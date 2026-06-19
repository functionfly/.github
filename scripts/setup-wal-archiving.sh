#!/usr/bin/env bash
#
# PostgreSQL WAL Archiving Setup Script
#
# This script configures PostgreSQL for WAL archiving to enable point-in-time recovery.
# It generates the necessary configuration and archive commands for R2/S3 storage.
#
# Usage:
#   ./scripts/setup-wal-archiving.sh --wal-dir=s3://bucket/wal --checkpoint=always
#
# Environment variables:
#   R2_ACCOUNT_ID      - Cloudflare R2 account ID
#   R2_ACCESS_KEY_ID   - R2 access key
#   R2_SECRET_ACCESS_KEY - R2 secret key
#   PGDATA             - PostgreSQL data directory (default: /var/lib/postgresql/data)

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

WAL_DIR=""
CHECKPOINT_MODE="fast"
BACKUP_USER="postgres"
OUTPUT_FILE=""

usage() {
    echo "PostgreSQL WAL Archiving Setup Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --wal-dir DIR          WAL archive destination (s3://bucket/wal or /path/to/wal)"
    echo "  --checkpoint MODE      Checkpoint mode: fast|spread (default: fast)"
    echo "  --backup-user USER     PostgreSQL user for backups (default: postgres)"
    echo "  --output FILE          Output config file (default: stdout)"
    echo "  --help                 Show this help"
    echo ""
    echo "Environment variables:"
    echo "  R2_ACCOUNT_ID          Cloudflare R2 account ID"
    echo "  R2_ACCESS_KEY_ID      R2 access key"
    echo "  R2_SECRET_ACCESS_KEY  R2 secret key"
    echo "  PGDATA                PostgreSQL data directory"
    echo ""
    exit 1
}

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --wal-dir) WAL_DIR="$2"; shift 2 ;;
        --checkpoint) CHECKPOINT_MODE="$2"; shift 2 ;;
        --backup-user) BACKUP_USER="$2"; shift 2 ;;
        --output) OUTPUT_FILE="$2"; shift 2 ;;
        --help) usage ;;
        *) log_error "Unknown option: $1"; usage ;;
    esac
done

if [ -z "$WAL_DIR" ]; then
    log_error "--wal-dir is required"
    usage
fi

# Determine if using R2/S3 or local storage
USE_CLOUD=false
if [[ "$WAL_DIR" == s3://* ]] || [[ "$WAL_DIR" == r2://* ]]; then
    USE_CLOUD=true
fi

# Check for cloud credentials if using cloud storage
if [ "$USE_CLOUD" = true ]; then
    if [ -z "${R2_ACCOUNT_ID:-}" ]; then
        log_error "R2_ACCOUNT_ID environment variable not set"
        exit 1
    fi
    if [ -z "${R2_ACCESS_KEY_ID:-}" ]; then
        log_error "R2_ACCESS_KEY_ID environment variable not set"
        exit 1
    fi
    if [ -z "${R2_SECRET_ACCESS_KEY:-}" ]; then
        log_error "R2_SECRET_ACCESS_KEY environment variable not set"
        exit 1
    fi
fi

# Get PostgreSQL data directory
PGDATA="${PGDATA:-/var/lib/postgresql/data}"
PGVERSION=$(psql --version | grep -oP '\d+' | head -1)

log_info "Configuring WAL archiving for PostgreSQL $PGVERSION"
log_info "WAL directory: $WAL_DIR"
log_info "Checkpoint mode: $CHECKPOINT_MODE"

# Build configuration
CONFIG=""

if [ "$USE_CLOUD" = true ]; then
    # R2/S3 configuration
    CONFIG="# ============================================
# PostgreSQL WAL Archiving Configuration
# Generated: $(date -Iseconds)
# WAL Directory: $WAL_DIR
# ============================================

# WAL Archiving
archive_mode = on
archive_command = 'aws s3 cp %p ${WAL_DIR}/%f --endpoint-url=https://\${R2_ACCOUNT_ID}.r2.cloudflarestorage.com'
archive_timeout = 300

# WAL Settings for PITR
wal_level = replica
max_wal_senders = 10
max_replication_slots = 10
wal_keep_size = 1GB

# Checkpoint settings
checkpoint_timeout = ${CHECKPOINT_MODE}
"

    # Create the archive command script
    ARCHIVE_SCRIPT="#!/bin/bash
# WAL Archive Command for Cloudflare R2
# This script uploads WAL files to R2

R2_ACCOUNT_ID=\"${R2_ACCOUNT_ID}\"
R2_ACCESS_KEY_ID=\"${R2_ACCESS_KEY_ID}\"
R2_SECRET_ACCESS_KEY=\"${R2_SECRET_ACCESS_KEY}\"
WAL_DIR=\"${WAL_DIR}\"

# Upload WAL file
aws s3 cp \"\$1\" \"\${WAL_DIR}/\${2}\" \
    --endpoint-url=\"https://\${R2_ACCOUNT_ID}.r2.cloudflarestorage.com\" \
    --storage-class REDUCED_REDUNDANCY

exit \$?
"

else
    # Local directory configuration
    CONFIG="# ============================================
# PostgreSQL WAL Archiving Configuration
# Generated: $(date -Iseconds)
# WAL Directory: $WAL_DIR
# ============================================

# WAL Archiving
archive_mode = on
archive_command = 'cp %p ${WAL_DIR}/%f'
archive_timeout = 300

# WAL Settings for PITR
wal_level = replica
max_wal_senders = 10
max_replication_slots = 10
wal_keep_size = 1GB

# Checkpoint settings
checkpoint_timeout = ${CHECKPOINT_MODE}
"

    ARCHIVE_SCRIPT="#!/bin/bash
# WAL Archive Command for Local Storage
# This script copies WAL files to local archive

WAL_DIR=\"${WAL_DIR}\"
mkdir -p \"\${WAL_DIR}\"
cp \"\$1\" \"\${WAL_DIR}/\${2}\"
exit \$?
"
fi

# Recovery configuration (for PITR)
RECOVERY_CONFIG="# ============================================
# PostgreSQL Recovery Configuration (PITR)
# Copy this to recovery.conf or configure via postgresql.conf
# ============================================

# Restore command to fetch WAL files from archive
restore_command = 'aws s3 cp ${WAL_DIR}/%f %p --endpoint-url=https://\${R2_ACCOUNT_ID}.r2.cloudflarestorage.com'

# Target time for point-in-time recovery (uncomment and set as needed)
# recovery_target_time = '2024-01-15 14:30:00'

# Action to take after recovery
recovery_target_action = 'promote'

# Set to 'pause' to review data before continuing
# recovery_target_action = 'pause'
"

# Output configuration
if [ -n "$OUTPUT_FILE" ]; then
    log_info "Writing configuration to: $OUTPUT_FILE"
    {
        echo "$CONFIG"
        echo ""
        echo "# ============================================"
        echo "# Recovery Configuration (PITR)"
        echo "# ============================================"
        echo "$RECOVERY_CONFIG"
    } > "$OUTPUT_FILE"

    # Create archive command script
    ARCHIVE_SCRIPT_FILE="${OUTPUT_FILE}.archive"
    log_info "Writing archive script to: $ARCHIVE_SCRIPT_FILE"
    echo "$ARCHIVE_SCRIPT" > "$ARCHIVE_SCRIPT_FILE"
    chmod +x "$ARCHIVE_SCRIPT_FILE"

    log_success "Configuration written successfully"
    echo ""
    echo "Next steps:"
    echo "1. Add the above configuration to postgresql.conf"
    echo "2. Create a replication slot: SELECT pg_create_physical_replication_slot('backup_slot');"
    echo "3. Reload PostgreSQL: pg_ctl reload"
    echo "4. Verify archiving: SELECT pg_stat_archiver;"
else
    echo "$CONFIG"
    echo ""
    echo "# ============================================"
    echo "# Recovery Configuration (PITR)"
    echo "# ============================================"
    echo "$RECOVERY_CONFIG"
fi

echo ""
log_info "To enable WAL archiving immediately (requires PostgreSQL restart):"
echo ""
echo "  # 1. Add config to postgresql.conf"
echo "  # 2. Create replication slot:"
echo "     psql -c \"SELECT pg_create_physical_replication_slot('backup_slot');\""
echo "  # 3. Restart PostgreSQL"
echo "     pg_ctl restart -D $PGDATA"
echo ""
log_warn "WAL archiving requires archive_mode=on, which needs a PostgreSQL restart"
log_warn "Ensure sufficient disk space for WAL files before enabling"
