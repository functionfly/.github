#!/bin/bash
# Complete automated deployment of AI Service to Fly.io
# Uses values from .env files where available

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
AI_APP="functionfly-ai-service"
MAIN_APP="functionfly-control"

echo "================================================"
echo "AI Service Complete Deployment"
echo "================================================"
echo ""

# Function to extract value from .env file
get_env_value() {
    local file="$1"
    local key="$2"
    if [ -f "$file" ]; then
        grep -E "^${key}=" "$file" 2>/dev/null | cut -d'=' -f2- | tr -d "'\"" || echo ""
    else
        echo ""
    fi
}

echo "📋 Step 1: Gathering Configuration"
echo "------------------------------------"

# Try to get Redis URL from .env files
REDIS_URL=$(get_env_value "$ROOT_DIR/.env" "REDIS_URL")
UPSTASH_URL=$(get_env_value "$ROOT_DIR/.env" "UPSTASH_REDIS_REST_URL")
UPSTASH_TOKEN=$(get_env_value "$ROOT_DIR/.env" "UPSTASH_REDIS_REST_TOKEN")

# Parse Upstash credentials from REDIS_URL if available
if [ -n "$REDIS_URL" ] && [[ $REDIS_URL =~ @(.+): ]]; then
    UPSTASH_ENDPOINT="${BASH_REMATCH[1]}"
    if [ -n "$UPSTASH_ENDPOINT" ] && [ -z "$UPSTASH_URL" ]; then
        UPSTASH_URL="https://${UPSTASH_ENDPOINT}"
    fi
    # Extract token from URL if present
    if [[ $REDIS_URL =~ default:([^@]+)@ ]]; then
        UPSTASH_TOKEN="${UPSTASH_TOKEN:-${BASH_REMATCH[1]}}"
    fi
fi

# Database - need production values
DB_SOURCE=""
if [ -f "$ROOT_DIR/.env.production" ]; then
    DATABASE_URL=$(get_env_value "$ROOT_DIR/.env.production" "DATABASE_URL")
    if [ -z "$DATABASE_URL" ]; then
        DB_HOST=$(get_env_value "$ROOT_DIR/.env.production" "DB_HOST")
        DB_PORT=$(get_env_value "$ROOT_DIR/.env.production" "DB_PORT")
        DB_USER=$(get_env_value "$ROOT_DIR/.env.production" "DB_USER")
        DB_PASSWORD=$(get_env_value "$ROOT_DIR/.env.production" "DB_PASSWORD")
        DB_NAME=$(get_env_value "$ROOT_DIR/.env.production" "DB_NAME")
        DB_SSLMODE=$(get_env_value "$ROOT_DIR/.env.production" "DB_SSLMODE")
    fi
fi

# AI API Keys
OPENAI_KEY=$(get_env_value "$ROOT_DIR/.env" "OPENAI_API_KEY")
if [ -z "$OPENAI_KEY" ]; then
    OPENAI_KEY=$(get_env_value "$ROOT_DIR/.env.production" "OPENAI_API_KEY")
fi

ANTHROPIC_KEY=$(get_env_value "$ROOT_DIR/.env" "ANTHROPIC_API_KEY")
if [ -z "$ANTHROPIC_KEY" ]; then
    ANTHROPIC_KEY=$(get_env_value "$ROOT_DIR/.env.production" "ANTHROPIC_API_KEY")
fi

OPENROUTER_KEY=$(get_env_value "$ROOT_DIR/.env" "OPENROUTER_API_KEY")

echo "Configuration found:"
echo "  Redis URL:        $([ -n "$REDIS_URL" ] && echo "✓ found" || echo "✗ missing")"
echo "  Upstash URL:      $([ -n "$UPSTASH_URL" ] && echo "✓ found" || echo "✗ missing")"
echo "  Upstash Token:    $([ -n "$UPSTASH_TOKEN" ] && echo "✓ found" || echo "✗ missing")"
echo "  Database URL:     $([ -n "$DATABASE_URL" ] && echo "✓ found" || echo "✗ missing")"
echo "  OpenAI Key:       $([ -n "$OPENAI_KEY" ] && echo "✓ found" || echo "✗ missing")"
echo "  Anthropic Key:    $([ -n "$ANTHROPIC_KEY" ] && echo "✓ found" || echo "✗ missing")"
echo "  OpenRouter Key:   $([ -n "$OPENROUTER_KEY" ] && echo "✓ found (optional)" || echo "✗ missing (optional)")"
echo ""

# Check if we have enough to proceed
MISSING=()
[ -z "$REDIS_URL" ] && [ -z "$UPSTASH_URL" ] && MISSING+=("Redis/Upstash")
[ -z "$DATABASE_URL" ] && [ -z "$DB_HOST" ] && MISSING+=("Database")
[ -z "$OPENAI_KEY" ] && [ -z "$ANTHROPIC_KEY" ] && MISSING+=("AI API Key (OpenAI or Anthropic)")

if [ ${#MISSING[@]} -gt 0 ]; then
    echo "⚠️  Missing required configuration: ${MISSING[*]}"
    echo ""
    echo "Please provide the missing values:"
    echo ""
fi

# Prompt for missing Redis/Upstash
if [ -z "$REDIS_URL" ] && [ -z "$UPSTASH_URL" ]; then
    read -p "REDIS_URL (or Upstash Redis URL): " REDIS_URL
fi

if [ -z "$UPSTASH_URL" ] && [ -n "$REDIS_URL" ]; then
    # Extract Upstash endpoint from REDIS_URL
    if [[ $REDIS_URL =~ @(.+): ]]; then
        UPSTASH_URL="https://${BASH_REMATCH[1]}"
    fi
fi

if [ -z "$UPSTASH_TOKEN" ]; then
    read -p "UPSTASH_REDIS_REST_TOKEN: " UPSTASH_TOKEN
fi

# Prompt for missing Database
if [ -z "$DATABASE_URL" ] && [ -z "$DB_HOST" ]; then
    read -p "DATABASE_URL (Neon PostgreSQL): " DATABASE_URL
fi

# Prompt for missing API keys
if [ -z "$OPENAI_KEY" ] && [ -z "$ANTHROPIC_KEY" ]; then
    read -p "OPENAI_API_KEY: " OPENAI_KEY
    if [ -z "$OPENAI_KEY" ]; then
        read -p "ANTHROPIC_API_KEY: " ANTHROPIC_KEY
    fi
fi

# Ensure we have at least one AI key
if [ -z "$OPENAI_KEY" ] && [ -z "$ANTHROPIC_KEY" ]; then
    echo "❌ Error: At least one AI provider API key is required"
    exit 1
fi

echo ""
echo "🔐 Step 2: Setting Secrets on Fly.io"
echo "------------------------------------"

# Build secrets command
SECRETS_CMD=""

# Database
if [ -n "$DATABASE_URL" ]; then
    SECRETS_CMD="DATABASE_URL=$DATABASE_URL"
fi

# Redis (Upstash)
if [ -n "$REDIS_URL" ]; then
    if [ -n "$SECRETS_CMD" ]; then
        SECRETS_CMD="$SECRETS_CMD REDIS_URL=$REDIS_URL"
    else
        SECRETS_CMD="REDIS_URL=$REDIS_URL"
    fi
fi

if [ -n "$UPSTASH_URL" ]; then
    SECRETS_CMD="$SECRETS_CMD UPSTASH_REDIS_REST_URL=$UPSTASH_URL"
fi

if [ -n "$UPSTASH_TOKEN" ]; then
    SECRETS_CMD="$SECRETS_CMD UPSTASH_REDIS_REST_TOKEN=$UPSTASH_TOKEN"
fi

# AI Provider Keys
if [ -n "$OPENAI_KEY" ]; then
    SECRETS_CMD="$SECRETS_CMD OPENAI_API_KEY=$OPENAI_KEY"
fi

if [ -n "$ANTHROPIC_KEY" ]; then
    SECRETS_CMD="$SECRETS_CMD ANTHROPIC_API_KEY=$ANTHROPIC_KEY"
fi

if [ -n "$OPENROUTER_KEY" ]; then
    SECRETS_CMD="$SECRETS_CMD OPENROUTER_API_KEY=$OPENROUTER_KEY"
fi

# Generate orchestrator API key
ORCH_KEY=$(openssl rand -hex 32)
SECRETS_CMD="$SECRETS_CMD ORCHESTRATOR_API_KEY=$ORCH_KEY"

echo "Setting secrets on $AI_APP..."
echo "  - DATABASE_URL"
echo "  - REDIS_URL"
echo "  - UPSTASH_REDIS_REST_URL"
echo "  - UPSTASH_REDIS_REST_TOKEN"
echo "  - OPENAI_API_KEY"
echo "  - ANTHROPIC_API_KEY"
[ -n "$OPENROUTER_KEY" ] && echo "  - OPENROUTER_API_KEY"
echo "  - ORCHESTRATOR_API_KEY (generated)"
echo ""

# Set all secrets
flyctl secrets set $SECRETS_CMD --app "$AI_APP"

echo ""
echo "✅ Secrets configured!"
echo ""

echo "🏗️  Step 3: Deploying AI Service"
echo "------------------------------------"
cd "$SCRIPT_DIR"
flyctl deploy --app "$AI_APP" --remote-only

echo ""
echo "⏳ Step 4: Health Check"
echo "------------------------------------"
sleep 15

# Check health via proxy
flyctl proxy 18081:8081 --app "$AI_APP" &
PROXY_PID=$!
sleep 5

HEALTH_URL="http://localhost:18081/health"
ATTEMPTS=0
MAX_ATTEMPTS=5
HEALTHY=false

while [ $ATTEMPTS -lt $MAX_ATTEMPTS ]; do
    if curl -sf "$HEALTH_URL" &>/dev/null; then
        echo "✅ Health check passed!"
        HEALTHY=true
        break
    fi
    ATTEMPTS=$((ATTEMPTS + 1))
    if [ $ATTEMPTS -lt $MAX_ATTEMPTS ]; then
        echo "  Attempt $ATTEMPTS/$MAX_ATTEMPTS failed, retrying..."
        sleep 5
    fi
done

# Cleanup proxy
kill $PROXY_PID 2>/dev/null || true

if [ "$HEALTHY" != true ]; then
    echo "❌ Health check failed!"
    echo "   Check logs: flyctl logs --app $AI_APP"
    exit 1
fi

echo ""
echo "🔗 Step 5: Configuring Orchestrator"
echo "------------------------------------"
echo "Updating orchestrator to use AI service..."
echo ""

flyctl secrets set \
    AI_SERVICE_URL="http://functionfly-ai-service.internal:8081" \
    AI_SERVICE_API_KEY="$ORCH_KEY" \
    --app "$MAIN_APP"

echo ""
echo "================================================"
echo "✅ AI Service Deployment Complete!"
echo "================================================"
echo ""
echo "App:               $AI_APP"
echo "Region:            ord (primary)"
echo "Internal URL:      http://$AI_APP.internal:8081"
echo "Orchestrator:      $MAIN_APP"
echo ""
echo "View logs:         flyctl logs --app $AI_APP"
echo "Status:            flyctl status --app $AI_APP"
echo "SSH console:       flyctl ssh console --app $AI_APP"
echo ""
echo "Next: Restart orchestrator to pick up new config:"
echo "  flyctl restart --app $MAIN_APP"
