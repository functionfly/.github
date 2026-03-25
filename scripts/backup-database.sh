#!/bin/bash
# FunctionFly Database Backup Script
# This script creates backups of the PostgreSQL database

set -e

# Configuration
BACKUP_DIR="/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/functionfly_${TIMESTAMP}.sql.gz"
RETENTION_DAYS=30

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create backup directory if it doesn't exist
mkdir -p "${BACKUP_DIR}"

# Create backup
log_info "Creating database backup..."
pg_dump -h postgres -U functionfly_prod -d functionfly_prod | gzip > "${BACKUP_FILE}"

if [ $? -eq 0 ]; then
    log_info "Backup created successfully: ${BACKUP_FILE}"
else
    log_error "Backup failed"
    exit 1
fi

# Verify backup
log_info "Verifying backup..."
if gunzip -t "${BACKUP_FILE}"; then
    log_info "Backup verification successful"
else
    log_error "Backup verification failed"
    exit 1
fi

# Clean up old backups
log_info "Cleaning up old backups..."
find "${BACKUP_DIR}" -name "functionfly_*.sql.gz" -mtime +${RETENTION_DAYS} -delete

log_info "Backup completed successfully"
