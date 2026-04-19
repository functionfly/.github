#!/bin/bash
# scripts/validate-migrations.sh
# Validates migration files for duplicate versions and naming convention compliance
# Usage: ./validate-migrations.sh [migrations_directory]

set -e

MIGRATIONS_DIR="${1:-migrations}"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "ERROR: Migrations directory not found: $MIGRATIONS_DIR"
    exit 1
fi

echo "Validating migrations in: $MIGRATIONS_DIR"
echo ""

# Check for duplicate versions
DUPLICATES=$(ls "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | \
    awk -F'[_/]' '{print $NF}' | \
    sed 's/\.up\.sql$//' | \
    sort | \
    uniq -d || true)

if [ -n "$DUPLICATES" ]; then
    echo "ERROR: Duplicate migration versions found:"
    echo "$DUPLICATES"
    echo ""
    echo "Affected files:"
    for dup in $DUPLICATES; do
        ls -la "$MIGRATIONS_DIR"/*"$dup"*.sql 2>/dev/null || true
    done
    echo ""
    echo "Run the following to see details:"
    echo "  ls $MIGRATIONS_DIR | grep -E '^[0-9]+_' | sort | uniq -d"
    exit 1
fi

# Check for mixed naming conventions (optional warning)
SEQUENTIAL_COUNT=$(ls "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | \
    grep -cE '/[0-9]{6}[^0-9]' || echo "0")
TIMESTAMP_COUNT=$(ls "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null | \
    grep -cE '/20[0-9]{12}[^0-9]' || echo "0")

echo "Naming convention check:"
echo "  - 6-digit sequential migrations: $SEQUENTIAL_COUNT"
echo "  - 14-digit timestamp migrations: $TIMESTAMP_COUNT"

if [ "$SEQUENTIAL_COUNT" -gt 0 ] && [ "$TIMESTAMP_COUNT" -gt 0 ]; then
    echo ""
    echo "WARNING: Mixed naming conventions detected."
    echo "Recommendation: Use timestamp format (YYYYMMDDHHMMSS) for all new migrations."
fi

# Check for missing .down.sql files
UP_FILES=$(ls "$MIGRATIONS_DIR"/*.up.sql 2>/dev/null)
MISSING_DOWN=0

for up_file in $UP_FILES; do
    down_file="${up_file%.up.sql}.down.sql"
    if [ ! -f "$down_file" ]; then
        echo "WARNING: Missing down migration for: $(basename "$up_file")"
        MISSING_DOWN=$((MISSING_DOWN + 1))
    fi
done

if [ $MISSING_DOWN -eq 0 ]; then
    echo ""
    echo "All migrations have corresponding .down.sql files."
fi

echo ""
echo "Migration validation passed!"
