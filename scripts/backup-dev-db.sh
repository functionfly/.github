#!/bin/bash
# FunctionFly Development Database Backup Script
# Backs up the local development PostgreSQL database (functionfly)
# Usage: ./scripts/backup-dev-db.sh [--restore <backup_file>]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${SCRIPT_DIR}/../backups/dev"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/functionfly_dev_${TIMESTAMP}.sql.gz"
RETENTION_DAYS=14

# Database connection (dev defaults from .env)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-functionfly}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Show usage
usage() {
    cat <<EOF
FunctionFly Development Database Backup Script

Usage: $0 [OPTIONS]

Options:
    --restore <file>    Restore from a backup file
    --list             List available backups
    --clean            Remove all dev backups
    -h, --help         Show this help message

Examples:
    $0                    # Create a new backup
    $0 --restore backups/dev/functionfly_dev_20250520_120000.sql.gz
    $0 --list
    $0 --clean

Environment variables (override defaults):
    DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME

EOF
}

# Ensure backup directory exists
ensure_backup_dir() {
    mkdir -p "${BACKUP_DIR}"
}

# Check if PostgreSQL is running
check_postgres() {
    if ! pg_isready -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" >/dev/null 2>&1; then
        log_error "PostgreSQL is not running on ${DB_HOST}:${DB_PORT}"
        log_info "Start it with: sudo pg_ctlcluster 17 main start"
        exit 1
    fi
}

# Create backup
do_backup() {
    check_postgres
    ensure_backup_dir

    log_step "Creating development database backup..."
    log_info "Database: ${DB_NAME} on ${DB_HOST}:${DB_PORT}"
    log_info "Output: ${BACKUP_FILE}"

    PGPASSWORD="${DB_PASSWORD}" pg_dump \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        --no-owner \
        --no-acl \
        | gzip > "${BACKUP_FILE}"

    if [ $? -eq 0 ]; then
        local size=$(du -h "${BACKUP_FILE}" | cut -f1)
        log_info "Backup created successfully: ${BACKUP_FILE} (${size})"
    else
        log_error "Backup failed"
        exit 1
    fi

    # Verify backup
    log_step "Verifying backup integrity..."
    if gunzip -t "${BACKUP_FILE}" 2>/dev/null; then
        log_info "Backup verification successful"
    else
        log_error "Backup verification failed - file may be corrupted"
        exit 1
    fi

    # Clean up old backups
    log_step "Cleaning up backups older than ${RETENTION_DAYS} days..."
    local deleted=$(find "${BACKUP_DIR}" -name "functionfly_dev_*.sql.gz" -mtime +${RETENTION_DAYS} -print -delete 2>/dev/null | wc -l)
    if [ "${deleted}" -gt 0 ]; then
        log_info "Removed ${deleted} old backup(s)"
    else
        log_info "No old backups to clean up"
    fi

    log_info "Backup completed successfully"
}

# List available backups
do_list() {
    ensure_backup_dir
    log_step "Available development backups in ${BACKUP_DIR}:"
    echo ""

    local backups=$(find "${BACKUP_DIR}" -name "functionfly_dev_*.sql.gz" -printf "%T@ %p\n" 2>/dev/null | sort -rn)
    
    if [ -z "${backups}" ]; then
        log_warn "No backups found"
        return
    fi

    printf "%-40s %15s\n" "BACKUP FILE" "SIZE"
    printf "%-40s %15s\n" "----------" "----"

    echo "${backups}" | while read timestamp file; do
        local size=$(du -h "${file}" 2>/dev/null | cut -f1)
        local name=$(basename "${file}")
        local date=$(date -d "@${timestamp}" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || echo "${timestamp}")
        printf "%-40s %15s\n" "${name}" "${size}"
    done
    echo ""
    log_info "Total backups: $(echo "${backups}" | wc -l)"
}

# Restore from backup
do_restore() {
    local restore_file="$1"

    if [ -z "${restore_file}" ]; then
        log_error "Restore file not specified"
        echo "Use: $0 --restore <backup_file>"
        echo ""
        log_info "Available backups:"
        do_list
        exit 1
    fi

    if [ ! -f "${restore_file}" ]; then
        log_error "Backup file not found: ${restore_file}"
        exit 1
    fi

    log_warn "This will overwrite the current development database!"
    log_warn "Database: ${DB_NAME} on ${DB_HOST}:${DB_PORT}"
    echo ""

    check_postgres

    read -p "Are you sure you want to restore? Type 'yes' to confirm: " confirm
    if [ "${confirm}" != "yes" ]; then
        log_info "Restore cancelled"
        exit 0
    fi

    # Create a pre-restore backup of current state
    log_step "Creating pre-restore backup of current state..."
    local pre_backup="${BACKUP_DIR}/pre_restore_$(date +%Y%m%d_%H%M%S).sql.gz"
    PGPASSWORD="${DB_PASSWORD}" pg_dump \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        --no-owner \
        --no-acl \
        | gzip > "${pre_backup}"
    log_info "Pre-restore backup saved: ${pre_backup}"

    log_step "Restoring database from ${restore_file}..."
    
    # Drop existing connections and restore
    PGPASSWORD="${DB_PASSWORD}" psql \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d postgres \
        -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${DB_NAME}' AND pid <> pg_backend_pid();" \
        >/dev/null 2>&1 || true

    PGPASSWORD="${DB_PASSWORD}" dropdb \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        --if-exists \
        "${DB_NAME}"

    PGPASSWORD="${DB_PASSWORD}" createdb \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        "${DB_NAME}"

    gunzip -c "${restore_file}" | PGPASSWORD="${DB_PASSWORD}" psql \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        >/dev/null

    if [ $? -eq 0 ]; then
        log_info "Database restored successfully from ${restore_file}"
    else
        log_error "Restore failed"
        log_info "Your pre-restore backup is available at: ${pre_backup}"
        exit 1
    fi
}

# Clean all dev backups
do_clean() {
    log_warn "This will delete ALL development database backups!"
    echo ""

    read -p "Are you sure you want to delete all backups? Type 'yes' to confirm: " confirm
    if [ "${confirm}" != "yes" ]; then
        log_info "Clean cancelled"
        exit 0
    fi

    local deleted=$(find "${BACKUP_DIR}" -name "functionfly_dev_*.sql.gz" -print -delete 2>/dev/null | wc -l)
    log_info "Deleted ${deleted} backup file(s)"
}

# Main
case "${1:-}" in
    --restore|-r)
        do_restore "${2:-}"
        ;;
    --list|-l)
        do_list
        ;;
    --clean|-c)
        do_clean
        ;;
    --help|-h)
        usage
        exit 0
        ;;
    "")
        do_backup
        ;;
    *)
        log_error "Unknown option: $1"
        usage
        exit 1
        ;;
esac
