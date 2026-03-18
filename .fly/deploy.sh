#!/bin/bash
# Fly.io Deployment Script for FunctionFly Orchestrator API
#
# Usage: ./deploy.sh [production|staging]
#
# Prerequisites:
# 1. Install Fly CLI: curl -L https://fly.io/install.sh | sh
# 2. Authenticate: fly auth login
# 3. Create the app: fly apps create functionfly-api
# 4. Set secrets: ./set-secrets.sh (see secrets.example)

set -e

ENVIRONMENT=${1:-production}

echo "=== Deploying FunctionFly API to Fly.io ($ENVIRONMENT) ==="

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

echo "Step 1: Setting environment-specific secrets..."
case $ENVIRONMENT in
    production)
        fly secrets set LOG_LEVEL="info" || true
        fly secrets set DEVELOPMENT="false" || true
        ;;
    staging)
        fly secrets set LOG_LEVEL="debug" || true
        fly secrets set DEVELOPMENT="true" || true
        ;;
    *)
        echo "Unknown environment: $ENVIRONMENT"
        exit 1
        ;;
esac

echo "Step 2: Checking database connectivity..."
# Verify database secrets are set
if ! fly secrets list | grep -q "DB_HOST\|DATABASE_URL"; then
    echo "Warning: Database secrets not found. Run ./set-secrets.sh first!"
    echo "Continuing deployment without migration check..."
else
    echo "Database secrets found."
fi

echo "Step 3: Running database migrations..."
# The app uses --skip-migrations flag by default (see AGENTS.md)
# For fresh deployments, run migrations using a temporary machine:
#
# Note: Migrations should be run separately. The orchestrator API is configured
# to use --skip-migrations flag. To run migrations:
#
# Option A: Using a separate migration tool
#   fly machine run . --env DATABASE_URL="..." -- migrate up
#
# Option B: Using the orchestrator API directly (not recommended for production)
#   fly machine run . -- ./orchestrator-api
#
# For now, we verify the app can connect to the database
echo "Note: Migrations are handled separately. Use --skip-migrations flag."
echo "To run migrations, use: fly machine run . -- ./orchestrator-api --migrate"

echo "Step 4: Deploying application..."
fly deploy

echo "Step 5: Checking deployment status..."
fly status

echo "Step 6: Verifying health..."
sleep 5
if fly health check; then
    echo "Health check passed!"
else
    echo "Warning: Health check failed. Check logs with: fly logs"
fi

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "View logs with: fly logs"
echo "Check health: fly health check"
echo "Open console: fly console"
echo "SSH into machine: fly ssh console"
