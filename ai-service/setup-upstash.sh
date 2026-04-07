#!/bin/bash
# Setup Upstash Redis for AI Service by copying from main app

set -euo pipefail

AI_APP="functionfly-ai-service"
MAIN_APP="functionfly-control"

echo "================================================"
echo "Setting Up Upstash Redis for AI Service"
echo "================================================"
echo ""

# Note: We can't read secret values directly (they're encrypted)
# But we know the keys exist from the previous check

echo "Copying Upstash Redis secrets from ${MAIN_APP} to ${AI_APP}..."
echo ""

# Since we can't read secrets, we need to get them from the user
# or from an env file
echo "The main app (${MAIN_APP}) has these Upstash secrets:"
echo "  - UPSTASH_REDIS_REST_URL"
echo "  - UPSTASH_REDIS_REST_TOKEN"
echo "  - REDIS_URL (standard Redis format)"
echo ""

echo "Please provide the Upstash Redis URL and Token."
echo "Get them from: https://console.upstash.com"
echo "Or copy from your .env or Infisical configuration."
echo ""

# Check if user has a .env file with these values
if [ -f ".env" ]; then
    echo "Found .env file. Checking for Upstash values..."
    UPSTASH_URL=$(grep -E "^UPSTASH_REDIS_REST_URL=" .env 2>/dev/null | cut -d'=' -f2- || echo "")
    UPSTASH_TOKEN=$(grep -E "^UPSTASH_REDIS_REST_TOKEN=" .env 2>/dev/null | cut -d'=' -f2- || echo "")
    REDIS_URL=$(grep -E "^REDIS_URL=" .env 2>/dev/null | cut -d'=' -f2- || echo "")

    if [ -n "$UPSTASH_URL" ] && [ -n "$UPSTASH_TOKEN" ]; then
        echo "✓ Found UPSTASH_REDIS_REST_URL and UPSTASH_REDIS_REST_TOKEN in .env"
        echo ""
        read -p "Use these values from .env? (Y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]] || [ -z "$REPLY" ]; then
            echo "Setting Upstash Redis secrets on ${AI_APP}..."
            flyctl secrets set \
                "UPSTASH_REDIS_REST_URL=$UPSTASH_URL" \
                "UPSTASH_REDIS_REST_TOKEN=$UPSTASH_TOKEN" \
                --app "$AI_APP"

            # Also set standard Redis URL for compatibility
            if [ -n "$REDIS_URL" ]; then
                flyctl secrets set "REDIS_URL=$REDIS_URL" --app "$AI_APP"
            fi

            echo ""
            echo "✅ Upstash Redis configured!"
            exit 0
        fi
    fi
fi

# Manual entry
echo "Please enter your Upstash Redis credentials:"
echo ""

read -p "UPSTASH_REDIS_REST_URL (e.g., https://xxx.upstash.io): " upstash_url
if [ -z "$upstash_url" ]; then
    echo "Error: URL is required"
    exit 1
fi

read -p "UPSTASH_REDIS_REST_TOKEN: " upstash_token
if [ -z "$upstash_token" ]; then
    echo "Error: Token is required"
    exit 1
fi

echo ""
echo "Setting secrets on ${AI_APP}..."
flyctl secrets set \
    "UPSTASH_REDIS_REST_URL=$upstash_url" \
    "UPSTASH_REDIS_REST_TOKEN=$upstash_token" \
    --app "$AI_APP"

# Also set the standard Redis URL format for libraries that expect it
# Upstash Redis URL format: rediss://default:token@endpoint:6379
if [[ $upstash_url =~ https://(.+) ]]; then
    ENDPOINT="${BASH_REMATCH[1]}"
    REDIS_URL="rediss://default:${upstash_token}@${ENDPOINT}:6379"
    echo ""
    echo "Setting REDIS_URL for compatibility..."
    flyctl secrets set "REDIS_URL=$REDIS_URL" --app "$AI_APP"
fi

echo ""
echo "✅ Upstash Redis configured for AI Service!"
echo ""
echo "Next: Set other required secrets and deploy:"
echo "  ./setup-secrets.sh  # for other secrets"
echo "  ./deploy.sh         # to deploy"
