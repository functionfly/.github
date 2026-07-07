#!/bin/bash
# Apply migrations to dedi server staging and production databases
# Usage: ./scripts/apply-dedi-migrations.sh [staging|prod|all]
# Default: all

set -e

SERVER="function@194.107.163.34"
MIGRATIONS_DIR="/opt/functionfly/migrations"
DB_PASSWORD='xNQhAgU6M7ubzyH6gjtglJHu0l9cX8P31IEf3lJEJvBIZpiTK2ue3tv7U+qjnvNt'

TARGET="${1:-all}"

echo "=========================================="
echo "Applying migrations to Dedi Server"
echo "Target: $TARGET"
echo "=========================================="

# Function to apply migrations
apply_migrations() {
    local dbname="$1"
    local name="$2"

    echo ""
    echo "----------------------------------------"
    echo "Checking $name ($dbname)..."
    echo "----------------------------------------"

    ssh -o StrictHostKeyChecking=no "$SERVER" "
        export PGPASSWORD='$DB_PASSWORD'
        cd $MIGRATIONS_DIR

        # Check current version
        echo 'Current migration version:'
        migrate -path . -database 'postgres://postgres@127.0.0.1:5432/$dbname?sslmode=require' version 2>&1 || echo 'No migrations yet'

        # Show pending
        echo ''
        echo 'Pending migrations would be applied...'

        # Apply migrations
        echo ''
        echo 'Applying migrations...'
        migrate -path . -database 'postgres://postgres@127.0.0.1:5432/$dbname?sslmode=require' up
    "

    echo "$name migrations complete!"
}

# Apply to staging
if [[ "$TARGET" == "staging" ]] || [[ "$TARGET" == "all" ]]; then
    apply_migrations "functionfly_staging" "STAGING"
fi

# Apply to production
if [[ "$TARGET" == "prod" ]] || [[ "$TARGET" == "all" ]]; then
    apply_migrations "functionfly_prod" "PRODUCTION"
fi

echo ""
echo "=========================================="
echo "All migrations complete!"
echo "=========================================="
