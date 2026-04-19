#!/bin/bash
# scripts/create-migration.sh
# Creates a new timestamped migration with proper naming convention
# Usage: ./scripts/create-migration.sh "description_of_changes"

set -e

MIGRATIONS_DIR="${MIGRATIONS_DIR:-migrations}"

if [ $# -lt 1 ]; then
    echo "Usage: $0 <description>"
    echo "Example: $0 'add_user_preferences_table'"
    echo ""
    echo "This creates:"
    echo "  migrations/YYYYMMDDHHMMSS_add_user_preferences_table.up.sql"
    echo "  migrations/YYYYMMDDHHMMSS_add_user_preferences_table.down.sql"
    exit 1
fi

# Get description and sanitize it (lowercase, replace spaces with underscores)
DESCRIPTION=$(echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')

# Generate timestamp
TIMESTAMP=$(date +%Y%m%d%H%M%S)

# Create filenames
UP_FILE="${MIGRATIONS_DIR}/${TIMESTAMP}_${DESCRIPTION}.up.sql"
DOWN_FILE="${MIGRATIONS_DIR}/${TIMESTAMP}_${DESCRIPTION}.down.sql"

# Check for potential duplicates (same timestamp - unlikely but possible)
if [ -f "$UP_FILE" ]; then
    echo "ERROR: Migration already exists: $UP_FILE"
    echo "Wait a minute and retry, or use a different description."
    exit 1
fi

# Create files with standard headers
cat > "$UP_FILE" << EOF
-- Migration: $DESCRIPTION
-- Created at: $(date -Iseconds)
-- Purpose: [TODO: Describe what this migration does]

-- Up migration
BEGIN;

-- TODO: Add your migration SQL here

COMMIT;
EOF

cat > "$DOWN_FILE" << EOF
-- Rollback: $DESCRIPTION
-- Created at: $(date -Iseconds)

-- Down migration (reverses the up migration)
BEGIN;

-- TODO: Add rollback SQL here
-- Example: DROP TABLE IF EXISTS new_table;

COMMIT;
EOF

echo "Created migration files:"
echo "  $UP_FILE"
echo "  $DOWN_FILE"
echo ""
echo "Next steps:"
echo "  1. Edit $UP_FILE to add your schema changes"
echo "  2. Edit $DOWN_FILE to add rollback logic"
echo "  3. Run ./scripts/validate-migrations.sh to check"
