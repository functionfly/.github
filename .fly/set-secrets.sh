#!/bin/bash
# Fly.io Secrets Setup Script for FunctionFly Orchestrator API
#
# Usage: ./set-secrets.sh [production|staging] [app-name]
#   Or:  FLY_APP=functionfly-control ./set-secrets.sh production
#
# If app-name (or FLY_APP) is not set, fly uses the app from fly.toml in the current directory.

set -e

ENVIRONMENT=${1:-production}
FLY_APP=${FLY_APP:-${2:-}}
APP_ARGS=""
if [ -n "$FLY_APP" ]; then
  APP_ARGS="--app $FLY_APP"
fi

echo "=== Setting Fly.io Secrets for FunctionFly API ($ENVIRONMENT) ==="
[ -n "$FLY_APP" ] && echo "App: $FLY_APP"

# Check if fly CLI is installed
if ! command -v fly &> /dev/null; then
    echo "Error: Fly CLI not installed. Run: curl -L https://fly.io/install.sh | sh"
    exit 1
fi

# Check if logged in
if ! fly auth whoami &> /dev/null; then
    echo "Error: Not logged in to Fly. Run: fly auth login"
    exit 1
fi

# =============================================================================
# CONFIGURATION - Edit these values before running!
# =============================================================================

# Neon PostgreSQL Configuration
# Option 1: Using DATABASE_URL (recommended - single variable)
# Get this from Neon Console > Connection Details > Connection string (pooled)
DATABASE_URL=""

# Option 2: Using individual DB_* variables (used if DATABASE_URL is empty)
DB_HOST=""
DB_PORT="5432"
DB_USER=""
DB_PASSWORD=""
DB_NAME="functionfly"
DB_SSLMODE="require"

# Redis Configuration
# Option A: Fly.io managed Redis (run: fly redis create)
REDIS_ADDR=""
REDIS_PASSWORD=""

# Option B: External Redis (e.g., Upstash, Redis Cloud)
# REDIS_ADDR="your-redis-host:6379"
# REDIS_PASSWORD="your-redis-password"

# Authentication & Security - Generate secure random values!
JWT_SECRET=""  # Minimum 32 characters
API_SHARED_SECRET=""  # Generate with: openssl rand -hex 32
DB_MASTER_KEY_PASSWORD=""  # Generate with: openssl rand -hex 32

# Application Configuration (required for coming-soon / CORS)
BASE_URL="https://api.functionfly.com"
CORS_ALLOWED_ORIGINS="https://functionfly.com,https://www.functionfly.com,https://app.functionfly.com,https://admin.functionfly.com"
FRONTEND_URL="https://functionfly.com"

# =============================================================================
# VALIDATION
# =============================================================================

echo "Validating configuration..."

# Check if DATABASE_URL or DB_HOST is set
if [ -z "$DATABASE_URL" ] && [ -z "$DB_HOST" ]; then
    echo "Error: Either DATABASE_URL or DB_HOST must be set"
    exit 1
fi

# Validate JWT_SECRET
if [ -n "$JWT_SECRET" ] && [ ${#JWT_SECRET} -lt 32 ]; then
    echo "Error: JWT_SECRET must be at least 32 characters"
    exit 1
fi

# =============================================================================
# SET SECRETS
# =============================================================================

echo "Setting secrets..."

# Database secrets
if [ -n "$DATABASE_URL" ]; then
    echo "Setting DATABASE_URL..."
    fly secrets set DATABASE_URL="$DATABASE_URL" $APP_ARGS
else
    echo "Setting DB_* variables..."
    [ -n "$DB_HOST" ] && fly secrets set DB_HOST="$DB_HOST" $APP_ARGS
    [ -n "$DB_PORT" ] && fly secrets set DB_PORT="$DB_PORT" $APP_ARGS
    [ -n "$DB_USER" ] && fly secrets set DB_USER="$DB_USER" $APP_ARGS
    [ -n "$DB_PASSWORD" ] && fly secrets set DB_PASSWORD="$DB_PASSWORD" $APP_ARGS
    [ -n "$DB_NAME" ] && fly secrets set DB_NAME="$DB_NAME" $APP_ARGS
    fly secrets set DB_SSLMODE="$DB_SSLMODE" $APP_ARGS
fi

# Redis secrets
if [ -n "$REDIS_ADDR" ]; then
    echo "Setting Redis secrets..."
    fly secrets set REDIS_ADDR="$REDIS_ADDR" $APP_ARGS
    [ -n "$REDIS_PASSWORD" ] && fly secrets set REDIS_PASSWORD="$REDIS_PASSWORD" $APP_ARGS
fi

# Authentication & Security
if [ -n "$JWT_SECRET" ]; then
    echo "Setting JWT_SECRET..."
    fly secrets set JWT_SECRET="$JWT_SECRET" $APP_ARGS
fi

if [ -n "$API_SHARED_SECRET" ]; then
    echo "Setting API_SHARED_SECRET..."
    fly secrets set API_SHARED_SECRET="$API_SHARED_SECRET" $APP_ARGS
fi

if [ -n "$DB_MASTER_KEY_PASSWORD" ]; then
    echo "Setting DB_MASTER_KEY_PASSWORD..."
    fly secrets set DB_MASTER_KEY_PASSWORD="$DB_MASTER_KEY_PASSWORD" $APP_ARGS
fi

# Application config
echo "Setting BASE_URL..."
fly secrets set BASE_URL="$BASE_URL" $APP_ARGS
[ -n "$CORS_ALLOWED_ORIGINS" ] && fly secrets set CORS_ALLOWED_ORIGINS="$CORS_ALLOWED_ORIGINS" $APP_ARGS
[ -n "$FRONTEND_URL" ] && fly secrets set FRONTEND_URL="$FRONTEND_URL" $APP_ARGS

# Environment-specific settings
case $ENVIRONMENT in
    production)
        echo "Setting production environment..."
        fly secrets set LOG_LEVEL="info" DEVELOPMENT="false" $APP_ARGS
        ;;
    staging)
        echo "Setting staging environment..."
        fly secrets set LOG_LEVEL="debug" DEVELOPMENT="true" $APP_ARGS
        ;;
esac

# =============================================================================
# VERIFY
# =============================================================================

echo ""
echo "Verifying secrets..."
fly secrets list $APP_ARGS

echo ""
echo "=== Secrets Setup Complete ==="
echo ""
echo "Next steps:"
echo "1. Run migrations (see docs/NEON.md)"
echo "2. Deploy the application: fly deploy"
echo "3. Check status: fly status"
echo "4. View logs: fly logs"
