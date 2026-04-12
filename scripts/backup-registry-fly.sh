#!/bin/bash
#
# FunctionFly Registry Backup Automation for Fly.io
#
# Performs automated backups of registry data to Cloudflare R2
# with cross-region replication to Backblaze B2.
#

set -euo pipefail

# Configuration
APP_NAME="functionfly-orchestrator"
BACKUP_NAME="registry-backup"
BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}
R2_BUCKET=${R2_BUCKET:-"functionfly-prod"}
B2_BUCKET=${B2_BUCKET:-"functionfly-backup-replica"}

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

# Load secrets from environment or Fly secrets
load_secrets() {
    # Try to get from environment first
    if [[ -z "${DATABASE_URL:-}" ]]; then
        log_info "Loading DATABASE_URL from Fly secrets..."
        DATABASE_URL=$(fly secrets get DATABASE_URL --app "$APP_NAME" 2>/dev/null | head -1)
    fi

    if [[ -z "${R2_ACCESS_KEY_ID:-}" ]]; then
        log_info "Loading R2 credentials from Fly secrets..."
        R2_ACCESS_KEY_ID=$(fly secrets get R2_ACCESS_KEY_ID --app "$APP_NAME" 2>/dev/null | head -1)
        R2_SECRET_ACCESS_KEY=$(fly secrets get R2_SECRET_ACCESS_KEY --app "$APP_NAME" 2>/dev/null | head -1)
        R2_ENDPOINT=$(fly secrets get R2_ENDPOINT --app "$APP_NAME" 2>/dev/null | head -1)
    fi

    # Check required vars
    local missing=()
    [[ -z "${DATABASE_URL:-}" ]] && missing+=("DATABASE_URL")
    [[ -z "${R2_ACCESS_KEY_ID:-}" ]] && missing+=("R2_ACCESS_KEY_ID")
    [[ -z "${R2_SECRET_ACCESS_KEY:-}" ]] && missing+=("R2_SECRET_ACCESS_KEY")
    [[ -z "${R2_ENDPOINT:-}" ]] && missing+=("R2_ENDPOINT")

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing required secrets: ${missing[*]}"
        exit 1
    fi
}

# Backup registry functions
backup_registry_functions() {
    log_info "Backing up registry functions..."

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
        log_success "Registry functions backup: $size"
        echo "$output_file"
    else
        log_error "Failed to backup registry functions"
        exit 1
    fi
}

# Backup registry metadata (JSON format for easier restoration)
backup_registry_metadata() {
    log_info "Exporting registry metadata..."

    local metadata_file="${TEMP_DIR}/${BACKUP_PREFIX}_metadata.json"

    # Export function list with metadata
    psql "$DATABASE_URL" -t -A -F"," -c "
        SELECT json_agg(row_to_json(t))
        FROM (
            SELECT
                rf.id,
                rf.author,
                rf.name,
                rf.latest_version,
                rf.title,
                rf.description,
                rf.category,
                rf.tags,
                rf.visibility,
                rf.price_per_call,
                rf.trust_score,
                rf.verified,
                rf.popularity_score,
                rf.reliability_score,
                rf.created_at,
                rf.updated_at,
                array_agg(DISTINCT rfv.version) as versions
            FROM registry_functions rf
            LEFT JOIN registry_function_versions rfv ON rf.id = rfv.function_id
            GROUP BY rf.id
        ) t;
    " > "$metadata_file"

    if [[ -s "$metadata_file" ]]; then
        log_success "Metadata export completed"
        echo "$metadata_file"
    else
        log_warn "Metadata export returned empty results"
        echo ""
    fi
}

# Upload to R2
upload_to_r2() {
    local file=$1
    local key=$2

    log_info "Uploading to R2: $key"

    # Use AWS CLI with R2 endpoint
    AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID" \
    AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY" \
    aws s3 cp "$file" "s3://${R2_BUCKET}/${key}" \
        --endpoint-url="$R2_ENDPOINT" \
        --storage-class STANDARD \
        2>/dev/null

    if [[ $? -eq 0 ]]; then
        log_success "Upload to R2 successful: $key"
        return 0
    else
        log_error "Failed to upload to R2: $key"
        return 1
    fi
}

# Replicate to B2 (optional)
replicate_to_b2() {
    local file=$1
    local key=$2

    # Check if B2 is configured
    if [[ -z "${B2_APPLICATION_KEY_ID:-}" || -z "${B2_APPLICATION_KEY:-}" ]]; then
        log_warn "B2 not configured, skipping cross-region replication"
        return 0
    fi

    log_info "Replicating to B2: $key"

    # Authorize and upload
    b2 authorize "$B2_APPLICATION_KEY_ID" "$B2_APPLICATION_KEY" &>/dev/null
    b2 upload-file "$B2_BUCKET" "$file" "$key" &>/dev/null

    if [[ $? -eq 0 ]]; then
        log_success "Replication to B2 successful"
        return 0
    else
        log_warn "Failed to replicate to B2 (non-critical)"
        return 0  # Don't fail backup for replica issues
    fi
}

# Create checksums
create_checksums() {
    local file=$1
    local checksum_file="${file}.sha256"

    sha256sum "$file" > "$checksum_file"
    log_info "Checksum: $(cat "$checksum_file")"
}

# Verify backup integrity
verify_backup() {
    local file=$1

    log_info "Verifying backup integrity..."

    # Check file is not empty
    if [[ ! -s "$file" ]]; then
        log_error "Backup file is empty"
        return 1
    fi

    # Verify gzip integrity if compressed
    if [[ "$file" == *.gz ]]; then
        if ! gzip -t "$file" 2>/dev/null; then
            log_error "Backup file is corrupted (gzip test failed)"
            return 1
        fi
    fi

    log_success "Backup integrity verified"
    return 0
}

# Cleanup old backups
cleanup_old_backups() {
    log_info "Cleaning up backups older than $BACKUP_RETENTION_DAYS days..."

    local cutoff_date=$(date -d "$BACKUP_RETENTION_DAYS days ago" +%Y%m%d 2>/dev/null || date -v-${BACKUP_RETENTION_DAYS}d +%Y%m%d)

    # List and delete old R2 backups
    AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID" \
    AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY" \
    aws s3 ls "s3://${R2_BUCKET}/" --endpoint-url="$R2_ENDPOINT" 2>/dev/null | \
    while read -r line; do
        local backup_date=$(echo "$line" | grep -o 'registry_[0-9]\{8\}' | head -1 | cut -d_ -f2)
        if [[ -n "$backup_date" && "$backup_date" < "$cutoff_date" ]]; then
            local key=$(echo "$line" | awk '{print $4}')
            log_info "Deleting old backup: $key"
            AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID" \
            AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY" \
            aws s3 rm "s3://${R2_BUCKET}/${key}" --endpoint-url="$R2_ENDPOINT" 2>/dev/null || true
        fi
    done

    log_success "Cleanup completed"
}

# Send notification
send_notification() {
    local status=$1
    local message=$2

    # Slack notification
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local payload=$(cat <<EOF
{
    "text": "Registry Backup - ${status}",
    "blocks": [
        {
            "type": "header",
            "text": {
                "type": "plain_text",
                "text": "💾 Registry Backup ${status}"
            }
        },
        {
            "type": "section",
            "fields": [
                {
                    "type": "mrkdwn",
                    "text": "*App:*\n${APP_NAME}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Time:*\n$(date)"
                }
            ]
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "${message}"
            }
        }
    ]
}
EOF
)

        curl -s -X POST -H 'Content-type: application/json' \
            --data "$payload" \
            "$SLACK_WEBHOOK_URL" &>/dev/null || true
    fi

    # Health check ping on success
    if [[ "$status" == "SUCCESS" && -n "${HEALTHCHECK_URL:-}" ]]; then
        curl -sf "$HEALTHCHECK_URL" &>/dev/null || true
    fi
}

# Main backup flow
main() {
    log_info "================================"
    log_info "FunctionFly Registry Backup"
    log_info "================================"
    log_info "Timestamp: $TIMESTAMP"

    local verify_only=false
    local list_backups=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --verify)
                verify_only=true
                shift
                ;;
            --list)
                list_backups=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --verify         Verify the most recent backup"
                echo "  --list           List available backups"
                echo "  --help           Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    # List mode
    if [[ "$list_backups" == "true" ]]; then
        log_info "Available backups in R2:"
        AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID" \
        AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY" \
        aws s3 ls "s3://${R2_BUCKET}/" --endpoint-url="$R2_ENDPOINT" 2>/dev/null | \
        grep "registry_" | sort -r | head -20
        exit 0
    fi

    # Load secrets
    load_secrets

    # Verify mode
    if [[ "$verify_only" == "true" ]]; then
        log_info "Verifying latest backup..."
        local latest_backup
        latest_backup=$(AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID" \
            AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY" \
            aws s3 ls "s3://${R2_BUCKET}/" --endpoint-url="$R2_ENDPOINT" 2>/dev/null | \
            grep "registry_.*functions.sql.gz" | sort -r | head -1 | awk '{print $4}')

        if [[ -n "$latest_backup" ]]; then
            log_info "Latest backup: $latest_backup"
            AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID" \
            AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY" \
            aws s3 cp "s3://${R2_BUCKET}/${latest_backup}" "${TEMP_DIR}/verify.sql.gz" \
                --endpoint-url="$R2_ENDPOINT" 2>/dev/null

            if verify_backup "${TEMP_DIR}/verify.sql.gz"; then
                log_success "Backup verification passed"
                exit 0
            else
                log_error "Backup verification failed"
                exit 1
            fi
        else
            log_error "No backups found"
            exit 1
        fi
    fi

    # Step 1: Backup registry functions
    local functions_backup
    functions_backup=$(backup_registry_functions)

    # Step 2: Verify backup
    if ! verify_backup "$functions_backup"; then
        log_error "Backup verification failed"
        send_notification "FAILED" "Backup verification failed - possible data corruption"
        exit 1
    fi

    # Step 3: Create checksum
    create_checksums "$functions_backup"

    # Step 4: Backup metadata
    local metadata_backup
    metadata_backup=$(backup_registry_metadata)

    # Step 5: Upload to R2
    local r2_key="backups/$(basename "$functions_backup")"
    if ! upload_to_r2 "$functions_backup" "$r2_key"; then
        send_notification "FAILED" "Failed to upload backup to R2"
        exit 1
    fi

    # Upload checksum
    upload_to_r2 "${functions_backup}.sha256" "${r2_key}.sha256"

    # Upload metadata if available
    if [[ -n "$metadata_backup" && -f "$metadata_backup" ]]; then
        upload_to_r2 "$metadata_backup" "backups/$(basename "$metadata_backup")"
    fi

    # Step 6: Cross-region replication
    replicate_to_b2 "$functions_backup" "$(basename "$functions_backup")"

    # Step 7: Cleanup old backups
    cleanup_old_backups

    # Success notification
    local backup_size=$(du -h "$functions_backup" | cut -f1)
    send_notification "SUCCESS" "Registry backup completed (${backup_size}). Uploaded to R2: ${r2_key}"

    log_success "================================"
    log_success "Backup Complete!"
    log_success "================================"
    log_info "Backup Size: $backup_size"
    log_info "R2 Location: s3://${R2_BUCKET}/${r2_key}"
    log_info "Retention: ${BACKUP_RETENTION_DAYS} days"

    return 0
}

# Run main function
main "$@"
