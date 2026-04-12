#!/bin/bash
# Fix trailing commas in functionfly.jsonc files

find functions/functionfly -name "functionfly.jsonc" -type f | while read -r file; do
    # Remove trailing commas before } or ]
    # Pattern: ,(\s*[}\]]) -> \1
    sed -i 's/,[[:space:]]*\([}\]]\)/\1/g' "$file"
done

echo "Fixed trailing commas in all manifests"
