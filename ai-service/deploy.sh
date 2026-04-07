#!/usr/bin/env bash
# Deploy script for FlyMind AI Service to Fly.io
# Usage: ./deploy.sh [production|staging]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_NAME="functionfly-ai-service"
ENVIRONMENT="${1:-production}"

echo "🚀 Deploying AI Service to Fly.io (${ENVIRONMENT})..."
echo "================================================"

# Verify we're in the right directory
if [ ! -f "${SCRIPT_DIR}/fly.toml" ]; then
    echo "❌ Error: fly.toml not found. Are you in the ai-service directory?"
    exit 1
fi

cd "$SCRIPT_DIR"

# Check if flyctl is installed
if ! command -v flyctl &> /dev/null; then
    echo "❌ Error: flyctl is not installed. Install it first:"
    echo "   curl -L https://fly.io/install.sh | sh"
    exit 1
fi

# Verify login
echo "🔐 Checking Fly.io authentication..."
if ! flyctl auth whoami &> /dev/null; then
    echo "❌ Not logged in to Fly.io. Run: flyctl auth login"
    exit 1
fi

# Check if app exists, create if not
echo "📦 Checking if app exists..."
if ! flyctl apps list | grep -q "^${APP_NAME}$"; then
    echo "   Creating app: ${APP_NAME}"
    flyctl apps create "$APP_NAME"
else
    echo "   App ${APP_NAME} exists"
fi

# Verify required secrets are set
echo "🔑 Checking required secrets..."
REQUIRED_SECRETS=(
    "DATABASE_URL"
    "REDIS_ADDR"
    "OPENAI_API_KEY"
    "ANTHROPIC_API_KEY"
    "ORCHESTRATOR_API_KEY"
)

MISSING_SECRETS=()
for secret in "${REQUIRED_SECRETS[@]}"; do
    if ! flyctl secrets list --app "$APP_NAME" 2>/dev/null | grep -q "^${secret}="; then
        MISSING_SECRETS+=("$secret")
    fi
done

if [ ${#MISSING_SECRETS[@]} -gt 0 ]; then
    echo "⚠️  Warning: Some secrets may not be set: ${MISSING_SECRETS[*]}"
    echo "   Set them with: flyctl secrets set SECRET_NAME=value --app ${APP_NAME}"
    echo ""
    read -p "Continue with deployment? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Deployment cancelled."
        exit 1
    fi
fi

# Build and deploy
echo "🏗️  Building and deploying..."
flyctl deploy --app "$APP_NAME" --remote-only

# Wait for deployment
echo "⏳ Waiting for deployment to stabilize..."
sleep 15

# Health check (using fly proxy)
echo "🩺 Running health check..."
flyctl proxy 18081:8081 --app "$APP_NAME" &
PROXY_PID=$!
sleep 5

HEALTH_URL="http://localhost:18081/health"
ATTEMPTS=0
MAX_ATTEMPTS=5

while [ $ATTEMPTS -lt $MAX_ATTEMPTS ]; do
    if curl -sf "$HEALTH_URL" &>/dev/null; then
        echo "✅ Health check passed!"
        break
    fi
    ATTEMPTS=$((ATTEMPTS + 1))
    if [ $ATTEMPTS -lt $MAX_ATTEMPTS ]; then
        echo "   Attempt $ATTEMPTS/$MAX_ATTEMPTS failed, retrying..."
        sleep 5
    fi
done

# Cleanup proxy
kill $PROXY_PID 2>/dev/null || true

if [ $ATTEMPTS -eq $MAX_ATTEMPTS ]; then
    echo "❌ Health check failed after $MAX_ATTEMPTS attempts!"
    echo "   Check logs: flyctl logs --app $APP_NAME"
    exit 1
fi

# Show status
echo ""
echo "✅ Deployment complete!"
echo "================================================"
echo "App:        $APP_NAME"
echo "Region:     ord (primary)"
echo "Internal:   http://$APP_NAME.internal:8081"
echo "Status:     $(flyctl status --app "$APP_NAME" | grep -E '(Status|Instances)')"
echo ""
echo "View logs:  flyctl logs --app $APP_NAME"
echo "SSH:        flyctl ssh console --app $APP_NAME"
