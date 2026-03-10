#!/bin/bash
# Script to fix duplicate migration numbers

cd migrations

# Get all migration files, sort them by modification time to preserve order
files=($(ls -t [0-9]*.sql | grep -E '^[0-9]+\_'))

# Create a temporary directory
mkdir -p temp_migrations

# Counter for new migration numbers
counter=1

# Process each file
for file in "${files[@]}"; do
    # Extract the migration name (everything after the first underscore)
    name=$(echo "$file" | sed 's/^[0-9]\+_//')
    
    # Create new filename with sequential number
    new_name=$(printf "%06d_%s" $counter "$name")
    
    # Copy to temp directory with new name
    cp "$file" "temp_migrations/$new_name"
    
    echo "Renamed $file -> $new_name"
    ((counter++))
done

# Replace the migrations directory
rm -f *.sql
mv temp_migrations/* .
rmdir temp_migrations

echo "Migration files renumbered successfully"
