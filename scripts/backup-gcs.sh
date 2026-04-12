#!/bin/bash
#
# Google Cloud Storage (GCS) Backup Script for FunctionFly
# Uses GCP credentials for backups instead of R2
#

set -euo pipefail

# Configuration
APP_NAME="functionfly-orchestrator"
BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-14}
GCS_BUCKET=${GCS_BUCKET:-"functionfly-backups"}

# Timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_PREFIX="registry_${TIMESTAMP}"

# Temporary directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Load secrets
load_secrets() {
    # Try to get from environment first
    if [[ -z "${DATABASE_URL:-}" ]]; then
        log_info "Loading DATABASE_URL from Fly secrets..."
        DATABASE_URL=$(fly secrets get DATABASE_URL --app "$APP_NAME" 2>/dev/null | head -1)
    fi

    # Check for GCS credentials
    if [[ -z "${GOOGLE_APPLICATION_CREDENTIALS:-}" ]]; then
        # Try to get from Fly secrets as base64 encoded
        log_info "Loading GCS credentials from Fly secrets..."
        local gcs_creds_b64
        gcs_creds_b64=$(fly secrets get GOOGLE_CREDENTIALS_B64 --app "$APP_NAME" 2>/dev/null | head -1)
        if [[ -n "$gcs_creds_b64" ]]; then
            echo "$gcs_creds_b64" | base64 -d > /tmp/gcs-creds.json
            export GOOGLE_APPLICATION_CREDENTIALS=/tmp/gcs-creds.json
            log_success "GCS credentials loaded"
        fi
    fi

    # Check required vars
    local missing=()
    [[ -z "${DATABASE_URL:-}" ]] && missing+=("DATABASE_URL")
    
    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing required secrets: ${missing[*]}"
        exit 1
    fi
}

# Backup registry functions to GCS
backup_to_gcs() {
    log_info "Backing up registry functions to GCS..."

    local output_file="${TEMP_DIR}/${BACKUP_PREFIX}_functions.sql.gz"

    # Selective backup of registry-related tables
    pg_dump \
        --no-owner \
        --no-acl \
        --clean \
        --if-exists \
        --table="registry_functions" \
        --table="registry_function_versions" \
        --table="registry_function_ratings" \
        --table="registry_function_signatures" \
        --table="registry_function_approvals" \
        --table="registry_function_verification_status" \
        --table="registry_function_malware_scans" \
        --table="registry_function_executions" \
        --table="registry_platform_fees" \
        --table="registry_user_wallets" \
        --table="execution_meg_records" \
        --table="execution_certificates" \
        --table="function_execution_passports" \
        --table="drift_reports" \
        --table="trust_history" \
        "$DATABASE_URL" | gzip > "$output_file"

    if [[ $? -eq 0 ]]; then
        local size=$(du -h "$output_file" | cut -f1)
        log_success "Backup created: $size"
    else
        log_error "Failed to create backup"
        exit 1
    fi

    # Upload to GCS
    if command -v gsutil &> /dev/null; then
        log_info "Uploading to GCS bucket: $GCS_BUCKET"
        gsutil -q cp "$output_file" "gs://${GCS_BUCKET}/backups/"
        
        # Create checksum
        sha256sum "$output_file" > "${output_file}.sha256"
        gsutil -q cp "${output_file}.sha256" "gs://${GCS_BUCKET}/backups/"
        
        log_success "Backup uploaded to GCS"
    else
        log_warn "gsutil not found, backup saved locally: $output_file"
    fi
}

# List GCS backups
list_backups() {
    log_info "Listing GCS backups..."
    if command -v gsutil &> /dev/null; then
        gsutil ls "gs://${GCS_BUCKET}/backups/" 2>/dev/null | sort -r | head -20
    else
        log_error "gsutil not installed"
    fi
}

# Cleanup old backups
cleanup_old_backups() {
    log_info "Cleaning up old backups (older than ${BACKUP_RETENTION_DAYS} days)..."
    
    if command -v gsutil &> /dev/null; then
        # GCS lifecycle policies handle this automatically
        # But we can also do manual cleanup
        log_info "GCS lifecycle policies should handle retention"
    fi
}

# Main
main() {
    log_info "================================"
    log_info "GCS Backup for FunctionFly"
    log_info "================================"

    case "${1:-backup}" in
        backup)
            load_secrets
            backup_to_gcs
            ;;
        list)
            list_backups
            ;;
        *)
            echo "Usage: $0 [backup|list]"
            exit 1
            ;;
    esac
}

main "$@"
