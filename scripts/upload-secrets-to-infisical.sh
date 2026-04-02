#!/bin/bash
# Upload secrets from .env to Infisical
# Usage: INFISICAL_TOKEN=... INFISICAL_PROJECT_ID=... ./scripts/upload-secrets-to-infisical.sh <env>

set -e

ENV="${1:-dev}"
PROJECT_ID="${INFISICAL_PROJECT_ID:-ef71805a-e65b-4f0b-9bc1-af1a8f39bb81}"
TOKEN="${INFISICAL_TOKEN:-}"

if [[ -z "$TOKEN" ]]; then
    echo "Error: INFISICAL_TOKEN not set"
    exit 1
fi

echo "Uploading secrets to Infisical (env=$ENV, project=$PROJECT_ID)"
echo ""

uploaded=0
skipped=0
failed=0

# Read .env file (skip comments, empty lines, shell exports, and inline comments)
while IFS= read -r line; do
    # Skip empty lines, comments, and shell export commands
    if [[ -z "$line" ]]; then
        continue
    fi
    if [[ "$line" == \#* ]]; then
        continue
    fi
    if [[ "$line" == export\ * ]]; then
        continue
    fi

    # Extract key and value (first = is the separator)
    key="${line%%=*}"
    value="${line#*=}"

    # Trim whitespace
    key="$(echo "$key" | xargs)"
    value="$(echo "$value" | xargs)"

    # Skip empty keys or shell export commands
    [[ -z "$key" || "$key" == "export" ]] && continue

    # Skip VITE_* and INFISICAL_* secrets (not allowed for Fly sync)
    if [[ "$key" == VITE_* ]] || [[ "$key" == INFISICAL_* ]]; then
        echo "  SKIP $key"
        ((skipped++)) || true
        continue
    fi

    # Skip tokens that look like Infisical tokens (avoid circular)
    if [[ "$key" == INFISICAL_TOKEN ]] || [[ "$key" == INFISICAL_PROJECT_ID ]]; then
        echo "  SKIP $key"
        ((skipped++)) || true
        continue
    fi

    # Set secret
    if infisical secrets set "$key=$value" --env="$ENV" --projectId="$PROJECT_ID" --token="$TOKEN" > /dev/null 2>&1; then
        echo "  SET $key"
        ((uploaded++)) || true
    else
        echo "  FAIL $key"
        ((failed++)) || true
    fi

done < .env

echo ""
echo "Done: $uploaded uploaded, $skipped skipped, $failed failed"
