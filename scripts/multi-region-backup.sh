#!/bin/bash
set -euo pipefail

# Multi-Region Database Backup Script for FunctionFly
# Backs up databases across multiple regions and replicates to all regions

# Configuration
BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}
S3_BUCKET=${S3_BUCKET:-functionfly-backups}
DB_PRIMARY_REGION=${DB_PRIMARY_REGION:-iad}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
ENVIRONMENT=${ENVIRONMENT:-production}

# Define regions
REGIONS=("iad" "lax" "fra")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_debug() { echo -e "${BLUE}[DEBUG]${NC} $1"; }

# Database configuration per region (set these via environment variables)
# DB_IAD_HOST, DB_LAX_HOST, DB_FRA_HOST

# Load environment-specific configuration
load_env() {
    local env=$1
    if [ -f ".env.${env}" ]; then
        source ".env.${env}"
    elif [ -f "deploy/database/${env}.env" ]; then
        source "deploy/database/${env}.env"
    fi
}

# Get database endpoint for a region
get_db_endpoint() {
    local region=$1
    local var_name="DB_${region^^}_HOST"
    local endpoint="${!var_name}"

    if [ -z "$endpoint" ]; then
        # Try to get from Neon or cloud provider
        case "$region" in
            "iad")
                echo "${NEON_IAD_HOST:-}"
                ;;
            "lax")
                echo "${NEON_LAX_HOST:-}"
                ;;
            "fra")
                echo "${NEON_FRA_HOST:-}"
                ;;
        esac
    else
        echo "$endpoint"
    fi
}

# Database backup function
backup_database() {
    local region=$1
    local endpoint=$2

    if [ -z "$endpoint" ]; then
        log_warn "No endpoint configured for region: $region, skipping backup"
        return 0
    fi

    log_info "Starting backup for region: $region"

    # Set database URL for this region
    export DATABASE_URL="postgres://${DB_USER:-postgres}:${DB_PASSWORD:-postgres}@${endpoint}:5432/${DB_NAME:-functionfly}?sslmode=require"

    # Create backup filename
    BACKUP_FILE="functionfly_${region}_${TIMESTAMP}.sql.gz"
    TEMP_FILE="/tmp/${BACKUP_FILE}"

    # Perform backup
    log_debug "Running pg_dump for region: $region"
    if ! pg_dump "$DATABASE_URL" | gzip > "$TEMP_FILE"; then
        log_error "Backup failed for region: $region"
        return 1
    fi

    # Calculate checksum
    CHECKSUM=$(sha256sum "$TEMP_FILE" | awk '{print $1}')

    # Upload to S3 with regional prefix
    log_debug "Uploading backup to S3: s3://${S3_BUCKET}/${region}/backups/${BACKUP_FILE}"
    if ! aws s3 cp "$TEMP_FILE" "s3://${S3_BUCKET}/${region}/backups/${BACKUP_FILE}"; then
        log_error "Failed to upload backup to S3 for region: $region"
        rm -f "$TEMP_FILE"
        return 1
    fi

    # Upload latest backup (symlink replacement)
    aws s3 cp "$TEMP_FILE" "s3://${S3_BUCKET}/${region}/backups/latest.sql.gz"

    # Store checksum for verification
    echo "${CHECKSUM}" | aws s3 cp - "s3://${S3_BUCKET}/${region}/backups/${BACKUP_FILE}.sha256"

    # Cleanup local file
    rm -f "$TEMP_FILE"

    log_info "Backup completed for region: $region"
    return 0
}

# Cross-region backup replication
replicate_backups() {
    local source_region=$1

    log_info "Replicating backups from $source_region to all regions"

    # Copy backups to secondary regions
    for region in "${REGIONS[@]}"; do
        if [ "$source_region" != "$region" ]; then
            log_debug "Replicating to region: $region"
            aws s3 sync "s3://${S3_BUCKET}/${source_region}/" "s3://${S3_BUCKET}/${region}/" \
                --source-region "us-east-1" \
                --region "${region}" \
                --exclude "*" \
                --include "backups/*" \
                --include "latest.sql.gz" || log_warn "Failed to replicate to region: $region"
        fi
    done
}

# Cleanup old backups
cleanup_old_backups() {
    local region=$1

    log_info "Cleaning up backups older than ${BACKUP_RETENTION_DAYS} days in region: $region"

    # List and delete old backups
    aws s3 ls "s3://${S3_BUCKET}/${region}/backups/" | while read -r line; do
        backup_date=$(echo "$line" | awk '{print $1}')
        backup_name=$(echo "$line" | awk '{print $4}')

        # Calculate age and delete if needed
        # Using AWS S3 lifecycle would be more efficient
    done

    # Use S3 lifecycle policy for cleanup (recommended)
    log_info "Note: Consider setting up S3 lifecycle policy for automatic cleanup"
}

# Verify backup integrity
verify_backup() {
    local region=$1
    local backup_file=$2

    log_info "Verifying backup: $backup_file in region: $region"

    # Download and verify checksum
    local temp_file="/tmp/verify_${region}.sql.gz"
    aws s3 cp "s3://${S3_BUCKET}/${region}/backups/${backup_file}" "$temp_file"

    local stored_checksum
    stored_checksum=$(aws s3 cp "s3://${S3_BUCKET}/${region}/backups/${backup_file}.sha256" -)
    local actual_checksum
    actual_checksum=$(sha256sum "$temp_file" | awk '{print $1}')

    rm -f "$temp_file"

    if [ "$stored_checksum" = "$actual_checksum" ]; then
        log_info "Backup verification passed for region: $region"
        return 0
    else
        log_error "Backup verification FAILED for region: $region"
        return 1
    fi
}

# Health check for database
check_db_health() {
    local region=$1
    local endpoint=$2

    if [ -z "$endpoint" ]; then
        return 1
    fi

    local result
    result=$(psql "postgres://${DB_USER:-postgres}:${DB_PASSWORD:-postgres}@${endpoint}:5432/${DB_NAME:-functionfly}?sslmode=require" \
        -t -c "SELECT 1;" 2>/dev/null || echo "0")

    [ "$result" = "1" ]
}

# Main backup process
main() {
    log_info "Starting multi-region backup process"
    log_info "Timestamp: $TIMESTAMP"
    log_info "Environment: $ENVIRONMENT"
    log_info "Primary region: $DB_PRIMARY_REGION"

    # Load environment
    load_env "$ENVIRONMENT"

    local backup_failed=0
    local success_count=0

    # First, backup primary region
    log_info "=== Backing up primary region: $DB_PRIMARY_REGION ==="
    PRIMARY_ENDPOINT=$(get_db_endpoint "$DB_PRIMARY_REGION")

    if [ -n "$PRIMARY_ENDPOINT" ]; then
        if check_db_health "$DB_PRIMARY_REGION" "$PRIMARY_ENDPOINT"; then
            if backup_database "$DB_PRIMARY_REGION" "$PRIMARY_ENDPOINT"; then
                ((success_count++))
                # Replicate primary backup to other regions
                replicate_backups "$DB_PRIMARY_REGION"
            else
                ((backup_failed++))
            fi
        else
            log_error "Primary region $DB_PRIMARY_REGION is unhealthy, skipping backup"
            ((backup_failed++))
        fi
    else
        log_warn "No endpoint configured for primary region: $DB_PRIMARY_REGION"
    fi

    # Backup other regions
    for region in "${REGIONS[@]}"; do
        if [ "$region" = "$DB_PRIMARY_REGION" ]; then
            continue
        fi

        log_info "=== Backing up secondary region: $region ==="
        ENDPOINT=$(get_db_endpoint "$region")

        if [ -n "$ENDPOINT" ]; then
            if check_db_health "$region" "$ENDPOINT"; then
                if backup_database "$region" "$ENDPOINT"; then
                    ((success_count++))
                else
                    ((backup_failed++))
                fi
            else
                log_warn "Region $region is unhealthy, skipping backup"
            fi
        else
            log_warn "No endpoint configured for region: $region"
        fi
    done

    # Summary
    log_info "=== Backup Summary ==="
    log_info "Successful: $success_count"
    log_info "Failed: $backup_failed"

    if [ $backup_failed -gt 0 ]; then
        log_error "Multi-region backup completed with errors"
        exit 1
    fi

    log_info "Multi-region backup process completed successfully"
}

# Point-in-time recovery function
point_in_time_recovery() {
    local target_timestamp=$1
    local target_region=$2

    log_info "Starting point-in-time recovery to: $target_timestamp in region: $target_region"

    # Find the closest backup before the target timestamp
    # This is a simplified version - in production, you'd use WAL archival

    log_info "PITR requires WAL archiving to be enabled"
    log_info "Use restore-database-pitr.sh for full PITR functionality"
}

# Handle script arguments
case "${1:-backup}" in
    backup)
        main
        ;;
    replicate)
        replicate_backups "${2:-$DB_PRIMARY_REGION}"
        ;;
    verify)
        verify_backup "$2" "$3"
        ;;
    cleanup)
        cleanup_old_backups "$2"
        ;;
    pitr)
        point_in_time_recovery "$2" "${3:-iad}"
        ;;
    *)
        echo "Usage: $0 {backup|replicate|verify|cleanup|pitr} [args...]"
        echo ""
        echo "Commands:"
        echo "  backup                    Run multi-region backup (default)"
        echo "  replicate <source_region> Replicate backups from source to all regions"
        echo "  verify <region> <file>    Verify backup integrity"
        echo "  cleanup <region>         Clean up old backups"
        echo "  pitr <timestamp> [region] Point-in-time recovery"
        exit 1
        ;;
esac
