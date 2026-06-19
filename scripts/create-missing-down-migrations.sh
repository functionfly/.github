#!/bin/bash
# scripts/create-missing-down-migrations.sh
# Creates stub down migrations for migrations that are missing them
# Usage: ./scripts/create-missing-down-migrations.sh [migrations_directory]

set -e

MIGRATIONS_DIR="${1:-migrations}"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "ERROR: Migrations directory not found: $MIGRATIONS_DIR"
    exit 1
fi

echo "Finding migrations missing .down.sql files in: $MIGRATIONS_DIR"
echo ""

CREATED=0
SKIPPED=0

for up_file in "$MIGRATIONS_DIR"/*.up.sql; do
    if [ ! -f "$up_file" ]; then
        continue
    fi

    down_file="${up_file%.up.sql}.down.sql"

    if [ ! -f "$down_file" ]; then
        filename=$(basename "$up_file")
        echo "Creating stub for: $filename"

        # Create a minimal down migration that does nothing but allows the migration system to work
        cat > "$down_file" << 'STUB_EOF'
-- Stub down migration
-- This is a placeholder for a migration that may not need reversal
-- or requires manual intervention for rollback

-- If this migration created tables/indexes that need explicit cleanup,
-- add the appropriate DROP statements below

STUB_EOF

        CREATED=$((CREATED + 1))
    else
        SKIPPED=$((SKIPPED + 1))
    fi
done

echo ""
echo "=== Summary ==="
echo "Down migrations created: $CREATED"
echo "Already existed: $SKIPPED"
echo ""

if [ $CREATED -gt 0 ]; then
    echo "WARNING: Review the created stub files and add proper rollback logic if needed."
    echo "Stub files only prevent the migration validator from complaining."
fi
