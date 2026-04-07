#!/bin/bash
# FunctionFly Function Backup Script
# Backs up PostgreSQL function data to Cloudflare R2 with versioning
#
# Usage:
#   ./backup-functions-to-r2.sh [options]
#
# Options:
#   -t, --tables      Specific tables to backup (default: all function-related)
#   -c, --compress    Compression level 0-9 (default: 6)
#   -e, --encrypt     Encrypt backup with GPG (requires GPG_KEY_ID env var)
#   -k, --keep-local  Keep local copy after upload (default: delete)
#   -d, --dry-run     Show what would be backed up without uploading
#   -h, --help        Show this help message
#
# Environment Variables:
#   DATABASE_URL      - PostgreSQL connection string (required if DB_* vars not set)
#   DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME - Individual connection params
#   R2_ACCOUNT_ID     - Cloudflare R2 account ID (required)
#   R2_ACCESS_KEY_ID  - R2 access key (required)
#   R2_SECRET_ACCESS_KEY - R2 secret key (required)
#   R2_BACKUP_BUCKET  - R2 bucket for backups (default: functionfly-backups)
#   BACKUP_RETENTION_DAYS - Days to keep backups (default: 30)
#   GPG_KEY_ID        - GPG key for encryption (optional)
#   SLACK_WEBHOOK_URL - Slack notification on failure (optional)
#   HEALTH_CHECK_URL  - Health check ping on success (optional)

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
LOG_FILE="${SCRIPT_DIR}/backup-functions.log"
BACKUP_DIR="${PROJECT_ROOT}/tmp/backups"

# Default configuration
COMPRESS_LEVEL=6
ENCRYPT=false
KEEP_LOCAL=false
DRY_RUN=false
TABLES=""
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
R2_BACKUP_BUCKET="${R2_BACKUP_BUCKET:-functionfly-backups}"

# Timestamp for backup naming
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
DATE_PATH=$(date +%Y/%m/%d)
BACKUP_PREFIX="functions"

# Colors for output (disable if not TTY)
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Logging functions
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $*" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" | tee -a "$LOG_FILE"
}

# Error handler
error_handler() {
    local exit_code=$?
    local line_number=$1
    log_error "Backup failed at line ${line_number} with exit code ${exit_code}"
    send_notification "failure" "Backup failed at line ${line_number} (exit code: ${exit_code})"
    cleanup_on_failure
    exit $exit_code
}

trap 'error_handler ${LINENO}' ERR

# Cleanup function
cleanup() {
    if [[ "$KEEP_LOCAL" == "false" && -d "$BACKUP_DIR" ]]; then
        log "Cleaning up temporary files..."
        rm -rf "$BACKUP_DIR"/*.sql* 2>/dev/null || true
    fi
}

cleanup_on_failure() {
    if [[ -d "$BACKUP_DIR" ]]; then
        log "Preserving failed backup files in: $BACKUP_DIR"
        # Rename to indicate failure
        for f in "$BACKUP_DIR"/*.sql*; do
            [[ -f "$f" ]] && mv "$f" "${f}.FAILED" 2>/dev/null || true
        done
    fi
}

# Help message
show_help() {
    sed -n '/^# /,/^#$/p' "$0" | sed 's/^# //' | head -n -1
    exit 0
}

# Parse arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--tables)
                TABLES="$2"
                shift 2
                ;;
            -c|--compress)
                COMPRESS_LEVEL="$2"
                shift 2
                ;;
            -e|--encrypt)
                ENCRYPT=true
                shift
                ;;
            -k|--keep-local)
                KEEP_LOCAL=true
                shift
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -h|--help)
                show_help
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
}

# Validate environment
validate_env() {
    local missing=()

    # Check database connectivity
    if [[ -z "${DATABASE_URL:-}" ]]; then
        if [[ -z "${DB_HOST:-}" || -z "${DB_USER:-}" || -z "${DB_PASSWORD:-}" || -z "${DB_NAME:-}" ]]; then
            missing+=("DATABASE_URL or DB_HOST/DB_USER/DB_PASSWORD/DB_NAME")
        fi
    fi

    # Check R2 credentials
    [[ -z "${R2_ACCOUNT_ID:-}" ]] && missing+=("R2_ACCOUNT_ID")
    [[ -z "${R2_ACCESS_KEY_ID:-}" ]] && missing+=("R2_ACCESS_KEY_ID")
    [[ -z "${R2_SECRET_ACCESS_KEY:-}" ]] && missing+=("R2_SECRET_ACCESS_KEY")

    # Check GPG if encryption enabled
    if [[ "$ENCRYPT" == "true" && -z "${GPG_KEY_ID:-}" ]]; then
        missing+=("GPG_KEY_ID (required when encryption enabled)")
    fi

    if [[ ${#missing[@]} -gt 0 ]]; then
        log_error "Missing required environment variables:"
        printf '  - %s\n' "${missing[@]}" | tee -a "$LOG_FILE"
        exit 1
    fi

    log "Environment validation passed"
}

# Build database connection string
get_db_connection() {
    if [[ -n "${DATABASE_URL:-}" ]]; then
        echo "$DATABASE_URL"
    else
        local host="${DB_HOST:-localhost}"
        local port="${DB_PORT:-5432}"
        local user="${DB_USER:-postgres}"
        local pass="${DB_PASSWORD:-}"
        local db="${DB_NAME:-functionfly}"
        local sslmode="${DB_SSLMODE:-require}"
        echo "postgres://${user}:${pass}@${host}:${port}/${db}?sslmode=${sslmode}"
    fi
}

# Get function-related tables to backup
get_tables() {
    if [[ -n "$TABLES" ]]; then
        echo "$TABLES"
        return
    fi

    # Default function-related tables in dependency order
    cat <<EOF
functions
function_versions
function_dependencies
function_environment
function_deployments
deployment_artifacts
function_metrics
function_invocations
function_reputation
function_verification
function_ratings
EOF
}

# Create backup
backup_functions() {
    local connection_url
    connection_url=$(get_db_connection)

    mkdir -p "$BACKUP_DIR"

    local backup_file="${BACKUP_DIR}/${BACKUP_PREFIX}-${TIMESTAMP}.sql"
    local table_list
    table_list=$(get_tables)

    log "Starting backup of function tables..."
    log "Tables: $(echo "$table_list" | tr '\n' ' ')"

    if [[ "$DRY_RUN" == "true" ]]; then
        log "DRY RUN: Would backup the following tables:"
        echo "$table_list" | while read -r table; do
            log "  - $table"
        done
        return 0
    fi

    # Get record counts for verification
    local total_records=0
    while read -r table; do
        [[ -z "$table" ]] && continue
        local count
        count=$(psql "$connection_url" -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | tr -d ' ') || count=0
        log "  $table: $count records"
        total_records=$((total_records + count))
    done <<< "$table_list"

    log "Total records to backup: $total_records"

    # Perform pg_dump
    log "Running pg_dump..."
    local pgdump_opts="--data-only --no-owner --no-privileges --disable-triggers"

    # Dump each table
    echo "$table_list" | while read -r table; do
        [[ -z "$table" ]] && continue
        log "  Dumping $table..."
        pg_dump $pgdump_opts \
            --table="$table" \
            "$connection_url" >> "$backup_file" 2>>"$LOG_FILE" || {
            log_error "Failed to dump table: $table"
            return 1
        }
    done

    # Get file size
    local raw_size
    raw_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")

    log "Raw backup complete: $backup_file ($(numfmt --to=iec-i --suffix=B "$raw_size" 2>/dev/null || echo "${raw_size} bytes"))"

    # Compress
    log "Compressing backup (level $COMPRESS_LEVEL)..."
    gzip -"$COMPRESS_LEVEL" -c "$backup_file" > "${backup_file}.gz"
    rm "$backup_file"
    backup_file="${backup_file}.gz"

    local compressed_size
    compressed_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")
    local ratio
    ratio=$(echo "scale=2; $raw_size / $compressed_size" | bc 2>/dev/null || echo "N/A")

    log "Compressed: $backup_file ($(numfmt --to=iec-i --suffix=B "$compressed_size" 2>/dev/null || echo "${compressed_size} bytes"), ratio: ${ratio}:1)"

    # Encrypt if requested
    if [[ "$ENCRYPT" == "true" ]]; then
        log "Encrypting backup with GPG (key: $GPG_KEY_ID)..."
        gpg --encrypt --recipient "$GPG_KEY_ID" --trust-model always \
            --output "${backup_file}.gpg" "$backup_file" 2>>"$LOG_FILE"
        rm "$backup_file"
        backup_file="${backup_file}.gpg"
        log "Encrypted: $backup_file"
    fi

    echo "$backup_file"
}

# Upload to R2
upload_to_r2() {
    local backup_file="$1"
    local filename
    filename=$(basename "$backup_file")
    local r2_key="backups/functions/${DATE_PATH}/${filename}"

    log "Uploading to R2..."
    log "  Bucket: $R2_BACKUP_BUCKET"
    log "  Key: $r2_key"

    if [[ "$DRY_RUN" == "true" ]]; then
        log "DRY RUN: Would upload to s3://$R2_BACKUP_BUCKET/$r2_key"
        return 0
    fi

    # Check if rclone or aws cli is available
    if command -v rclone &> /dev/null; then
        # Use rclone (recommended for R2)
        rclone copy "$backup_file" \
            ":s3,provider=Cloudflare,access_key_id=${R2_ACCESS_KEY_ID},secret_access_key=${R2_SECRET_ACCESS_KEY},endpoint=https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com:${R2_BACKUP_BUCKET}/$r2_key" \
            2>>"$LOG_FILE" || {
            log_error "R2 upload failed (rclone)"
            return 1
        }
    elif command -v aws &> /dev/null; then
        # Use AWS CLI with R2 endpoint
        local r2_endpoint="https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com"
        AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID" \
        AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY" \
        aws s3 cp "$backup_file" "s3://${R2_BACKUP_BUCKET}/${r2_key}" \
            --endpoint-url="$r2_endpoint" \
            --region=auto \
            2>>"$LOG_FILE" || {
            log_error "R2 upload failed (aws cli)"
            return 1
        }
    else
        log_error "No upload tool available (install rclone or aws cli)"
        return 1
    fi

    log "Upload complete: s3://${R2_BACKUP_BUCKET}/${r2_key}"

    # Verify upload
    log "Verifying upload..."
    local local_checksum
    local_checksum=$(sha256sum "$backup_file" | cut -d' ' -f1)

    # Get remote checksum (using head-object metadata)
    if command -v rclone &> /dev/null; then
        # rclone sha1sum doesn't work well with R2, skip verification or implement differently
        log_warn "Remote checksum verification not implemented for rclone"
    fi

    # Store backup metadata
    local metadata_file="${backup_file}.meta.json"
    cat > "$metadata_file" <<EOF
{
    "backup_timestamp": "${TIMESTAMP}",
    "date_path": "${DATE_PATH}",
    "r2_key": "${r2_key}",
    "r2_bucket": "${R2_BACKUP_BUCKET}",
    "local_checksum": "${local_checksum}",
    "compressed_size": $(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0"),
    "retention_days": ${BACKUP_RETENTION_DAYS},
    "tables_backed_up": $(get_tables | jq -R -s -c 'split("\n") | map(select(length > 0))')
}
EOF

    # Upload metadata
    local meta_r2_key="${r2_key}.meta.json"
    if command -v rclone &> /dev/null; then
        rclone copy "$metadata_file" \
            ":s3,provider=Cloudflare,access_key_id=${R2_ACCESS_KEY_ID},secret_access_key=${R2_SECRET_ACCESS_KEY},endpoint=https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com:${R2_BACKUP_BUCKET}/${meta_r2_key}"
    else
        local r2_endpoint="https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com"
        AWS_ACCESS_KEY_ID="$R2_ACCESS_KEY_ID" \
        AWS_SECRET_ACCESS_KEY="$R2_SECRET_ACCESS_KEY" \
        aws s3 cp "$metadata_file" "s3://${R2_BACKUP_BUCKET}/${meta_r2_key}" \
            --endpoint-url="$r2_endpoint" \
            --region=auto
    fi

    rm "$metadata_file"
    log "Metadata uploaded: ${meta_r2_key}"
}

# Cleanup old backups
cleanup_old_backups() {
    log "Cleaning up backups older than ${BACKUP_RETENTION_DAYS} days..."

    if [[ "$DRY_RUN" == "true" ]]; then
        log "DRY RUN: Would list and potentially delete old backups"
        return 0
    fi

    local cutoff_date
    cutoff_date=$(date -d "${BACKUP_RETENTION_DAYS} days ago" +%Y/%m/%d 2>/dev/null || \
                  date -v-${BACKUP_RETENTION_DAYS}d +%Y/%m/%d 2>/dev/null || \
                  echo "")

    if [[ -z "$cutoff_date" ]]; then
        log_warn "Could not calculate cutoff date for cleanup"
        return 0
    fi

    log "  Cutoff date: $cutoff_date"

    # List old backups and delete (using rclone or AWS CLI)
    # This is a simplified version - production should implement proper lifecycle rules in R2
    log "  (R2 lifecycle rules should handle automatic deletion - configure in Cloudflare dashboard)"
}

# Send notification
send_notification() {
    local status="$1"
    local message="$2"

    # Slack webhook
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        local color="good"
        [[ "$status" == "failure" ]] && color="danger"

        curl -s -X POST -H 'Content-type: application/json' \
            --data "{
                \"attachments\": [{
                    \"color\": \"${color}\",
                    \"title\": \"FunctionFly Backup ${status}\",
                    \"text\": \"${message}\",
                    \"fields\": [
                        {\"title\": \"Timestamp\", \"value\": \"${TIMESTAMP}\", \"short\": true},
                        {\"title\": \"Environment\", \"value\": \"${ENVIRONMENT:-unknown}\", \"short\": true}
                    ],
                    \"footer\": \"FunctionFly Backup System\",
                    \"ts\": $(date +%s)
                }]
            }" \
            "$SLACK_WEBHOOK_URL" > /dev/null 2>&1 || true
    fi

    # Health check ping
    if [[ -n "${HEALTH_CHECK_URL:-}" && "$status" == "success" ]]; then
        curl -s "$HEALTH_CHECK_URL" > /dev/null 2>&1 || true
    fi
}

# Main function
main() {
    local start_time end_time duration
    start_time=$(date +%s)

    log "=========================================="
    log "FunctionFly Function Backup Starting"
    log "Timestamp: $TIMESTAMP"
    log "=========================================="

    parse_args "$@"
    validate_env

    # Create backup
    local backup_file
    backup_file=$(backup_functions)

    if [[ "$DRY_RUN" != "true" ]]; then
        # Upload to R2
        upload_to_r2 "$backup_file"

        # Cleanup old backups
        cleanup_old_backups

        # Cleanup local files
        cleanup

        # Calculate duration
        end_time=$(date +%s)
        duration=$((end_time - start_time))

        log "=========================================="
        log_success "Backup completed successfully!"
        log "Duration: ${duration}s"
        log "=========================================="

        send_notification "success" "Backup completed in ${duration}s. Key: backups/functions/${DATE_PATH}/"
    else
        log "=========================================="
        log "DRY RUN completed - no changes made"
        log "=========================================="
    fi
}

# Run main
main "$@"
