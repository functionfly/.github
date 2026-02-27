#!/bin/bash

set -euo pipefail

# Database restore script for FunctionFly
# Usage: ./restore-database.sh [environment] [backup_file_or_latest]

ENVIRONMENT=${1:-production}
BACKUP_SPEC=${2:-latest}  # Can be 'latest' or specific filename

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

# Storage configuration (same as backup script)
STORAGE_BACKEND=${DB_BACKUP_STORAGE_BACKEND:-local}
S3_ENDPOINT=${DB_BACKUP_S3_ENDPOINT:-}
S3_BUCKET=${DB_BACKUP_S3_BUCKET:-}
S3_PREFIX=${DB_BACKUP_S3_PREFIX:-backups/}

# Determine backup file to restore
if [ "$BACKUP_SPEC" = "latest" ]; then
    echo "Finding latest backup..."
    BACKUP_FILE=$(find_backup_file)
else
    BACKUP_FILE="$BACKUP_SPEC"
fi

if [ -z "$BACKUP_FILE" ]; then
    echo "Error: No backup file found for '$BACKUP_SPEC'"
    exit 1
fi

echo "Restoring from: $BACKUP_FILE"
echo "Target database: $DB_NAME"
echo "⚠️  WARNING: This will REPLACE the current database!"
read -p "Are you sure you want to continue? (yes/no): " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Restore cancelled."
    exit 1
fi

# Download backup if needed
LOCAL_BACKUP_FILE="/tmp/functionfly_restore.backup"

download_backup() {
    local remote_file="$1"
    local local_file="$2"

    case "$STORAGE_BACKEND" in
        "local")
            # Assume file is already local
            if [ -f "$remote_file" ]; then
                cp "$remote_file" "$local_file"
            else
                echo "Error: Local backup file not found: $remote_file"
                exit 1
            fi
            ;;

        "s3"|"aws")
            echo "Downloading from AWS S3..."
            if [ -n "$S3_ENDPOINT" ]; then
                aws s3 cp --endpoint-url="$S3_ENDPOINT" "s3://${S3_BUCKET}/${remote_file}" "$local_file"
            else
                aws s3 cp "s3://${S3_BUCKET}/${remote_file}" "$local_file"
            fi
            ;;

        "b2"|"backblaze")
            echo "Downloading from Backblaze B2..."
            b2 download-file-by-name "$S3_BUCKET" "$remote_file" "$local_file"
            ;;

        "gcs"|"google")
            echo "Downloading from Google Cloud Storage..."
            gsutil cp "gs://${S3_BUCKET}/${remote_file}" "$local_file"
            ;;

        "wasabi")
            echo "Downloading from Wasabi..."
            aws s3 cp --endpoint-url="https://s3.wasabisys.com" "s3://${S3_BUCKET}/${remote_file}" "$local_file"
            ;;

        "minio")
            echo "Downloading from MinIO..."
            mc cp "minio/${S3_BUCKET}/${remote_file}" "$local_file"
            ;;

        "scp"|"sftp")
            echo "Downloading via SCP..."
            local scp_opts="-o StrictHostKeyChecking=no"
            if [ -n "$SCP_KEY_FILE" ]; then
                scp_opts="$scp_opts -i $SCP_KEY_FILE"
            fi
            scp $scp_opts "${SCP_USER}@${SCP_HOST}:${SCP_PATH}/${remote_file}" "$local_file"
            ;;

        "rsync")
            echo "Downloading via rsync..."
            rsync -avz "$RSYNC_DESTINATION/${remote_file}" "$local_file"
            ;;

        *)
            echo "Error: Unknown storage backend '$STORAGE_BACKEND'"
            exit 1
            ;;
    esac
}

find_backup_file() {
    case "$STORAGE_BACKEND" in
        "local")
            # Find latest local backup
            find /var/backups/functionfly -name "functionfly_${ENVIRONMENT}_*.sql.gz" -type f -printf '%T@ %p\n' | sort -n | tail -1 | cut -d' ' -f2-
            ;;

        "s3"|"aws"|"wasabi")
            # List and find latest from S3
            local s3_cmd="aws s3 ls s3://${S3_BUCKET}/${S3_PREFIX}functionfly_${ENVIRONMENT}_"
            if [ -n "$S3_ENDPOINT" ]; then
                s3_cmd="$s3_cmd --endpoint-url=$S3_ENDPOINT"
            fi
            $s3_cmd | sort | tail -1 | awk '{print $4}'
            ;;

        "b2"|"backblaze")
            # List and find latest from B2
            b2 ls "$S3_BUCKET" "${S3_PREFIX}functionfly_${ENVIRONMENT}_" | sort | tail -1 | awk '{print $2}'
            ;;

        "gcs"|"google")
            # List and find latest from GCS
            gsutil ls "gs://${S3_BUCKET}/${S3_PREFIX}functionfly_${ENVIRONMENT}_*" | sort | tail -1 | xargs basename
            ;;

        "scp"|"sftp")
            # SSH to remote and find latest
            ssh -o StrictHostKeyChecking=no ${SCP_KEY_FILE:+-i $SCP_KEY_FILE} "${SCP_USER}@${SCP_HOST}" \
                "ls -t ${SCP_PATH}/functionfly_${ENVIRONMENT}_*.sql.gz | head -1 | xargs basename"
            ;;

        "rsync")
            # Check local rsync destination
            ls -t "$RSYNC_DESTINATION"/functionfly_${ENVIRONMENT}_*.sql.gz 2>/dev/null | head -1 | xargs basename
            ;;

        *)
            echo ""
            ;;
    esac
}

# Download the backup
echo "Downloading backup..."
download_backup "$BACKUP_FILE" "$LOCAL_BACKUP_FILE"

# Verify download
if [ ! -f "$LOCAL_BACKUP_FILE" ]; then
    echo "Error: Failed to download backup file"
    exit 1
fi

echo "Backup downloaded successfully: $(du -h "$LOCAL_BACKUP_FILE")"

# Create backup of current database (safety measure)
CURRENT_BACKUP="/tmp/pre_restore_$(date +%Y%m%d_%H%M%S).sql"
echo "Creating safety backup of current database..."
PGPASSWORD="$DB_PASSWORD" pg_dump \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    --format=custom \
    --file="$CURRENT_BACKUP"

echo "Safety backup created: $CURRENT_BACKUP"

# Terminate active connections to the database
echo "Terminating active connections to database..."
TERMINATE_QUERY="
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();
"

PGPASSWORD="$DB_PASSWORD" psql \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="postgres" \
    --command="$TERMINATE_QUERY" 2>/dev/null || true

# Drop and recreate the database
echo "Recreating database..."
PGPASSWORD="$DB_PASSWORD" psql \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="postgres" \
    --command="DROP DATABASE IF EXISTS $DB_NAME;"

PGPASSWORD="$DB_PASSWORD" psql \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="postgres" \
    --command="CREATE DATABASE $DB_NAME;"

# Restore from backup
echo "Restoring database from backup..."
PGPASSWORD="$DB_PASSWORD" pg_restore \
    --host="$DB_HOST" \
    --port="$DB_PORT" \
    --username="$DB_USER" \
    --dbname="$DB_NAME" \
    --verbose \
    --clean \
    --if-exists \
    --create \
    "$LOCAL_BACKUP_FILE"

# Run post-restore migrations if needed
echo "Running post-restore setup..."
# Add any post-restore setup here

# Clean up
echo "Cleaning up temporary files..."
rm -f "$LOCAL_BACKUP_FILE"

echo "✅ Database restore completed successfully!"
echo ""
echo "📋 Restore Summary:"
echo "  - Restored from: $BACKUP_FILE"
echo "  - Target database: $DB_NAME"
echo "  - Safety backup: $CURRENT_BACKUP"
echo ""
echo "⚠️  Important: Test your application thoroughly after restore!"
echo "   The safety backup is at: $CURRENT_BACKUP"