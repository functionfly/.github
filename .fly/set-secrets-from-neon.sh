#!/bin/bash
# Set Fly.io secrets using Neon CLI for DATABASE_URL and generated auth secrets.
#
# Usage: ./set-secrets-from-neon.sh [production|staging]
#   Or:  FLY_APP=functionfly-control ./set-secrets-from-neon.sh production
#
# Requires: neon CLI (make neon-install), fly CLI, neon auth (or NEON_API_KEY).
# Optional: set REDIS_ADDR, REDIS_PASSWORD, JWT_SECRET, API_SHARED_SECRET, DB_MASTER_KEY_PASSWORD
#   before running to use existing values; otherwise JWT/API/DB_MASTER are generated.

set -e

ENVIRONMENT=${1:-production}
FLY_APP=${FLY_APP:-functionfly-control}
APP_ARGS="--app $FLY_APP"
NEON_BRANCH=${NEON_BRANCH:-production}

echo "=== Setting Fly.io secrets from Neon (branch: $NEON_BRANCH, app: $FLY_APP) ==="

if ! command -v fly &> /dev/null; then
    echo "Error: Fly CLI not installed. Run: curl -L https://fly.io/install.sh | sh"
    exit 1
fi
if ! fly auth whoami &> /dev/null; then
    echo "Error: Not logged in to Fly. Run: fly auth login"
    exit 1
fi
if ! command -v neon &> /dev/null; then
    echo "Error: Neon CLI not installed. Run: make neon-install && neon auth"
    exit 1
fi

# DATABASE_URL from Neon (pooled, functionfly database)
echo "Getting Neon connection string (pooled, database: functionfly)..."
DATABASE_URL=$(neon connection-string "$NEON_BRANCH" --pooled --database-name functionfly 2>/dev/null) || true
if [ -z "$DATABASE_URL" ]; then
    echo "Error: Could not get Neon connection string. Run: neon auth"
    exit 1
fi

# Generate secrets if not set
if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(openssl rand -hex 32)
    echo "Generated JWT_SECRET"
fi
if [ -z "$API_SHARED_SECRET" ]; then
    API_SHARED_SECRET=$(openssl rand -hex 32)
    echo "Generated API_SHARED_SECRET"
fi
if [ -z "$DB_MASTER_KEY_PASSWORD" ]; then
    DB_MASTER_KEY_PASSWORD=$(openssl rand -hex 32)
    echo "Generated DB_MASTER_KEY_PASSWORD"
fi

# Application URLs (required for API and for coming-soon page: CORS allows functionfly.com to POST /v1/feedback)
BASE_URL="${BASE_URL:-https://api.functionfly.com}"
CORS_ALLOWED_ORIGINS="${CORS_ALLOWED_ORIGINS:-https://functionfly.com,https://www.functionfly.com}"
FRONTEND_URL="${FRONTEND_URL:-https://functionfly.com}"

echo "Setting secrets..."

fly secrets set DATABASE_URL="$DATABASE_URL" $APP_ARGS
fly secrets set JWT_SECRET="$JWT_SECRET" $APP_ARGS
fly secrets set API_SHARED_SECRET="$API_SHARED_SECRET" $APP_ARGS
fly secrets set DB_MASTER_KEY_PASSWORD="$DB_MASTER_KEY_PASSWORD" $APP_ARGS
fly secrets set BASE_URL="$BASE_URL" $APP_ARGS
fly secrets set CORS_ALLOWED_ORIGINS="$CORS_ALLOWED_ORIGINS" $APP_ARGS
fly secrets set FRONTEND_URL="$FRONTEND_URL" $APP_ARGS

if [ -n "$REDIS_ADDR" ]; then
    echo "Setting Redis..."
    fly secrets set REDIS_ADDR="$REDIS_ADDR" $APP_ARGS
    [ -n "$REDIS_PASSWORD" ] && fly secrets set REDIS_PASSWORD="$REDIS_PASSWORD" $APP_ARGS
else
    echo "Skipping REDIS_ADDR (not set). Create with: fly redis create --name functionfly-control-redis"
fi

case $ENVIRONMENT in
    production)
        fly secrets set LOG_LEVEL="info" DEVELOPMENT="false" $APP_ARGS
        ;;
    staging)
        fly secrets set LOG_LEVEL="debug" DEVELOPMENT="true" $APP_ARGS
        ;;
esac

echo ""
fly secrets list $APP_ARGS
echo ""
echo "=== Done ==="
echo "Coming-soon page: CORS_ALLOWED_ORIGINS and FRONTEND_URL are set so functionfly.com can POST to /v1/feedback."
echo "If you need Redis: fly redis create --name functionfly-control-redis, then set REDIS_ADDR and re-run."
