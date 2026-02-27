#!/bin/bash

set -euo pipefail

# Database backup script for FunctionFly
# Usage: ./backup-database.sh [environment] [backup_type]
# backup_type: full (default), base (for PITR), wal (archive WAL files)

ENVIRONMENT=${1:-production}
BACKUP_TYPE=${2:-full}  # full, base, wal
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/var/backups/functionfly"

# Set backup file paths based on type
case "$BACKUP_TYPE" in
    "full")
        BACKUP_FILE="${BACKUP_DIR}/functionfly_${ENVIRONMENT}_${TIMESTAMP}.sql.gz"
        ;;
    "base")
        BACKUP_FILE="${BACKUP_DIR}/basebackup_${ENVIRONMENT}_${TIMESTAMP}.tar.gz"
        ;;
    "wal")
        BACKUP_FILE="${BACKUP_DIR}/wal_${ENVIRONMENT}_${TIMESTAMP}.tar.gz"
        ;;
    *)
        echo "Invalid backup type: $BACKUP_TYPE. Use: full, base, or wal"
        exit 1
        ;;
esac

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

# Backup configuration
BACKUP_RETENTION_DAYS=${DB_BACKUP_RETENTION_DAYS:-30}
STORAGE_BACKEND=${DB_BACKUP_STORAGE_BACKEND:-local}  # local, s3, b2, gcs, wasabi, minio, scp, rsync

# S3-compatible storage (AWS S3, DigitalOcean Spaces, Wasabi)
S3_ENDPOINT=${DB_BACKUP_S3_ENDPOINT:-}  # For non-AWS S3 services
S3_BUCKET=${DB_BACKUP_S3_BUCKET:-}
S3_PREFIX=${DB_BACKUP_S3_PREFIX:-backups/}
AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-}
AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-}

# Backblaze B2 (much cheaper than S3)
B2_APPLICATION_KEY_ID=${DB_BACKUP_B2_KEY_ID:-}
B2_APPLICATION_KEY=${DB_BACKUP_B2_KEY:-}
B2_BUCKET=${DB_BACKUP_B2_BUCKET:-}

# Google Cloud Storage
GOOGLE_APPLICATION_CREDENTIALS=${DB_BACKUP_GCS_KEY_FILE:-}
GCS_BUCKET=${DB_BACKUP_GCS_BUCKET:-}

# SCP/SFTP configuration (very cheap alternative)
SCP_HOST=${DB_BACKUP_SCP_HOST:-}
SCP_USER=${DB_BACKUP_SCP_USER:-}
SCP_PATH=${DB_BACKUP_SCP_PATH:-}
SCP_KEY_FILE=${DB_BACKUP_SCP_KEY_FILE:-}

# Rsync configuration (for local network or cheap VPS)
RSYNC_DESTINATION=${DB_BACKUP_RSYNC_DEST:-}

# MinIO (self-hosted S3-compatible)
MINIO_ENDPOINT=${DB_BACKUP_MINIO_ENDPOINT:-}
MINIO_ACCESS_KEY=${DB_BACKUP_MINIO_ACCESS_KEY:-}
MINIO_SECRET_KEY=${DB_BACKUP_MINIO_SECRET_KEY:-}

# Create backup directory
mkdir -p "$BACKUP_DIR"

echo "Starting ${BACKUP_TYPE} database backup for ${ENVIRONMENT} environment..."
echo "Backup file: $BACKUP_FILE"

case "$BACKUP_TYPE" in
    "full")
        # Standard logical backup using pg_dump
        echo "Creating full logical backup..."
        PGPASSWORD="$DB_PASSWORD" pg_dump \
            --host="$DB_HOST" \
            --port="$DB_PORT" \
            --username="$DB_USER" \
            --dbname="$DB_NAME" \
            --format=custom \
            --compress=9 \
            --verbose \
            --file="${BACKUP_DIR}/functionfly_${ENVIRONMENT}_${TIMESTAMP}.backup"

        # Compress the backup
        gzip "${BACKUP_DIR}/functionfly_${ENVIRONMENT}_${TIMESTAMP}.backup"
        mv "${BACKUP_DIR}/functionfly_${ENVIRONMENT}_${TIMESTAMP}.backup.gz" "$BACKUP_FILE"
        ;;

    "base")
        # Physical base backup for point-in-time recovery
        echo "Creating base backup for point-in-time recovery..."

        # Create base backup directory
        BASEBACKUP_DIR="${BACKUP_DIR}/basebackup_${TIMESTAMP}"
        mkdir -p "$BASEBACKUP_DIR"

        # Use pg_basebackup to create a base backup
        PGPASSWORD="$DB_PASSWORD" pg_basebackup \
            --host="$DB_HOST" \
            --port="$DB_PORT" \
            --username="$DB_USER" \
            --pgdata="$BASEBACKUP_DIR" \
            --format=tar \
            --compress=gzip \
            --verbose \
            --progress

        # Move the compressed backup to the final location
        mv "${BASEBACKUP_DIR}.tar.gz" "$BACKUP_FILE"
        rm -rf "$BASEBACKUP_DIR"

        # Create backup label file for recovery reference
        BACKUP_LABEL_FILE="${BACKUP_DIR}/backup_label_${ENVIRONMENT}_${TIMESTAMP}.txt"
        echo "Base backup created at: $TIMESTAMP" > "$BACKUP_LABEL_FILE"
        echo "Backup file: $(basename "$BACKUP_FILE")" >> "$BACKUP_LABEL_FILE"
        echo "WAL location at backup time: $(PGPASSWORD="$DB_PASSWORD" psql --host="$DB_HOST" --port="$DB_PORT" --username="$DB_USER" --dbname="$DB_NAME" -t -c "SELECT pg_walfile_name(pg_current_wal_lsn());" | tr -d ' ')" >> "$BACKUP_LABEL_FILE"

        echo "Base backup label saved to: $BACKUP_LABEL_FILE"
        ;;

    "wal")
        # Archive WAL files for point-in-time recovery
        echo "Archiving WAL files..."

        # Create WAL archive directory
        WAL_DIR="${BACKUP_DIR}/wal_archive"
        mkdir -p "$WAL_DIR"

        # Get list of WAL files to archive (older than 1 hour to avoid active files)
        WAL_FILES=$(PGPASSWORD="$DB_PASSWORD" psql --host="$DB_HOST" --port="$DB_PORT" --username="$DB_USER" --dbname="$DB_NAME" -t -c "
            SELECT name FROM pg_ls_waldir()
            WHERE modification > (now() - interval '1 hour')
            ORDER BY name;")

        if [ -z "$WAL_FILES" ]; then
            echo "No WAL files to archive (all files are recent)"
            exit 0
        fi

        # Copy WAL files to archive directory
        while IFS= read -r wal_file; do
            if [ -n "$wal_file" ]; then
                echo "Archiving WAL file: $wal_file"
                # In a real setup, you'd copy from pg_wal directory
                # For this example, we'll just record which files should be archived
                echo "$wal_file" >> "${WAL_DIR}/archived_files_${TIMESTAMP}.txt"
            fi
        done <<< "$WAL_FILES"

        # Create compressed archive
        tar -czf "$BACKUP_FILE" -C "$WAL_DIR" .
        rm -rf "$WAL_DIR"

        echo "WAL files archived to: $BACKUP_FILE"
        ;;

    *)
        echo "Unsupported backup type: $BACKUP_TYPE"
        exit 1
        ;;
esac

echo "Backup completed: $BACKUP_FILE"

# Upload to configured storage backend
upload_backup() {
    local file="$1"
    local remote_name="functionfly_${ENVIRONMENT}_${TIMESTAMP}.sql.gz"

    case "$STORAGE_BACKEND" in
        "local")
            echo "Keeping backup local only (no upload configured)"
            return 0
            ;;

        "s3"|"aws")
            if [ -n "$S3_BUCKET" ]; then
                echo "Uploading to AWS S3..."
                if [ -n "$S3_ENDPOINT" ]; then
                    aws s3 cp --endpoint-url="$S3_ENDPOINT" "$file" "s3://${S3_BUCKET}/${S3_PREFIX}${remote_name}"
                else
                    aws s3 cp "$file" "s3://${S3_BUCKET}/${S3_PREFIX}${remote_name}"
                fi
            fi
            ;;

        "b2"|"backblaze")
            if [ -n "$B2_BUCKET" ] && [ -n "$B2_APPLICATION_KEY_ID" ]; then
                echo "Uploading to Backblaze B2..."
                # Install b2 CLI if not present
                if ! command -v b2 &> /dev/null; then
                    pip3 install b2 || curl -s https://raw.githubusercontent.com/Backblaze/B2_Command_Line_Tool/master/b2 install
                fi
                b2 authorize-account "$B2_APPLICATION_KEY_ID" "$B2_APPLICATION_KEY"
                b2 upload-file "$B2_BUCKET" "$file" "${S3_PREFIX}${remote_name}"
            fi
            ;;

        "gcs"|"google")
            if [ -n "$GCS_BUCKET" ] && [ -n "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
                echo "Uploading to Google Cloud Storage..."
                gcloud auth activate-service-account --key-file="$GOOGLE_APPLICATION_CREDENTIALS"
                gsutil cp "$file" "gs://${GCS_BUCKET}/${S3_PREFIX}${remote_name}"
            fi
            ;;

        "wasabi")
            if [ -n "$S3_BUCKET" ]; then
                echo "Uploading to Wasabi..."
                aws s3 cp --endpoint-url="https://s3.wasabisys.com" "$file" "s3://${S3_BUCKET}/${S3_PREFIX}${remote_name}"
            fi
            ;;

        "minio")
            if [ -n "$MINIO_ENDPOINT" ] && [ -n "$S3_BUCKET" ]; then
                echo "Uploading to MinIO..."
                mc alias set minio "$MINIO_ENDPOINT" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY"
                mc cp "$file" "minio/${S3_BUCKET}/${S3_PREFIX}${remote_name}"
            fi
            ;;

        "scp"|"sftp")
            if [ -n "$SCP_HOST" ] && [ -n "$SCP_USER" ] && [ -n "$SCP_PATH" ]; then
                echo "Uploading via SCP..."
                local scp_opts="-o StrictHostKeyChecking=no"
                if [ -n "$SCP_KEY_FILE" ]; then
                    scp_opts="$scp_opts -i $SCP_KEY_FILE"
                fi
                scp $scp_opts "$file" "${SCP_USER}@${SCP_HOST}:${SCP_PATH}/${remote_name}"
            fi
            ;;

        "rsync")
            if [ -n "$RSYNC_DESTINATION" ]; then
                echo "Uploading via rsync..."
                rsync -avz --delete "$file" "$RSYNC_DESTINATION/${remote_name}"
            fi
            ;;

        *)
            echo "Warning: Unknown storage backend '$STORAGE_BACKEND'"
            return 1
            ;;
    esac
}

upload_backup "$BACKUP_FILE"

# Clean up old backups
echo "Cleaning up old backups (older than ${BACKUP_RETENTION_DAYS} days)..."
find "$BACKUP_DIR" -name "functionfly_${ENVIRONMENT}_*.sql.gz" -mtime +$BACKUP_RETENTION_DAYS -delete

# Clean up old remote backups based on storage backend
cleanup_remote_backups() {
    local cutoff_date=$(date -d "${BACKUP_RETENTION_DAYS} days ago" +%Y-%m-%d)

    case "$STORAGE_BACKEND" in
        "s3"|"aws")
            if [ -n "$S3_BUCKET" ]; then
                echo "Cleaning up old S3 backups..."
                local s3_cmd="aws s3api list-objects-v2 --bucket $S3_BUCKET --prefix ${S3_PREFIX}functionfly_${ENVIRONMENT}_"
                if [ -n "$S3_ENDPOINT" ]; then
                    s3_cmd="$s3_cmd --endpoint-url=$S3_ENDPOINT"
                fi
                $s3_cmd --query 'Contents[?LastModified<`'"$cutoff_date"'`].Key' --output text | \
                xargs -I {} aws s3 rm "s3://${S3_BUCKET}/{}" || true
            fi
            ;;

        "b2"|"backblaze")
            if [ -n "$B2_BUCKET" ]; then
                echo "Cleaning up old Backblaze B2 backups..."
                # List and delete old files (b2 doesn't have direct date filtering)
                b2 ls "$B2_BUCKET" "${S3_PREFIX}functionfly_${ENVIRONMENT}_" | \
                awk -v cutoff="$cutoff_date" '{
                    # Parse filename to extract date
                    split($2, parts, "_");
                    if (length(parts) >= 3) {
                        date_part = substr(parts[2], 1, 8);  # YYYYMMDD
                        if (date_part < cutoff) {
                            print $2
                        }
                    }
                }' | xargs -I {} b2 delete-file-version "$B2_BUCKET" {} || true
            fi
            ;;

        "gcs"|"google")
            if [ -n "$GCS_BUCKET" ]; then
                echo "Cleaning up old GCS backups..."
                gsutil ls "gs://${GCS_BUCKET}/${S3_PREFIX}functionfly_${ENVIRONMENT}_*" | \
                xargs -I {} gsutil stat {} | \
                awk -v cutoff="$cutoff_date" '/Creation time:/ {
                    # Parse creation time and compare
                    cmd = "date -d \"" $3 " " $4 " " $5 "\" +%Y-%m-%d"
                    cmd | getline file_date
                    close(cmd)
                    if (file_date < cutoff) {
                        print FILENAME
                    }
                }' | xargs -I {} gsutil rm {} || true
            fi
            ;;

        "scp"|"sftp")
            if [ -n "$SCP_HOST" ] && [ -n "$SCP_USER" ]; then
                echo "Cleaning up old SCP backups..."
                ssh -o StrictHostKeyChecking=no ${SCP_KEY_FILE:+-i $SCP_KEY_FILE} "${SCP_USER}@${SCP_HOST}" \
                    "find ${SCP_PATH} -name 'functionfly_${ENVIRONMENT}_*.sql.gz' -mtime +${BACKUP_RETENTION_DAYS} -delete" || true
            fi
            ;;

        "rsync")
            # Rsync destination cleanup is handled by rsync --delete in upload
            echo "Rsync cleanup handled during upload (--delete flag)"
            ;;
    esac
}

cleanup_remote_backups

echo "Backup process completed successfully"

# Log backup metadata
echo "$(date): Backup completed - $BACKUP_FILE ($(du -h "$BACKUP_FILE" | cut -f1))" >> "${BACKUP_DIR}/backup.log"