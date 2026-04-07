#!/usr/bin/env bash
# Setup secrets for AI Service by copying from main app and adding AI-specific keys

set -euo pipefail

AI_APP="functionfly-ai-service"
MAIN_APP="functionfly-control"

echo "================================================"
echo "Setting up secrets for AI Service"
echo "================================================"
echo ""
echo "This will copy shared secrets from ${MAIN_APP} and set AI-specific ones."
echo ""

# Check if user has required API keys
echo "Required AI Service API Keys:"
echo "  1. OPENAI_API_KEY      (from https://platform.openai.com/api-keys)"
echo "  2. ANTHROPIC_API_KEY   (from https://console.anthropic.com/settings/keys)"
echo "  3. ORCHESTRATOR_API_KEY (we'll generate a secure one)"
echo ""

# Check existing secrets on main app
echo "Checking secrets on ${MAIN_APP}..."
flyctl secrets list --app "$MAIN_APP" | grep -E "(DATABASE_URL|REDIS_ADDR|REDIS_PASSWORD)" || true
echo ""

echo "You need to manually copy these secrets from ${MAIN_APP} to ${AI_APP}."
echo "Run these commands with your actual values:"
echo ""
echo "# Copy database and Redis from main app:"
echo "flyctl secrets set DATABASE_URL='postgresql://...' --app ${AI_APP}"
echo "flyctl secrets set REDIS_ADDR='...' --app ${AI_APP}"
echo "flyctl secrets set REDIS_PASSWORD='...' --app ${AI_APP}"
echo ""
echo "# AI Provider API Keys:"
echo "flyctl secrets set OPENAI_API_KEY='sk-...' --app ${AI_APP}"
echo "flyctl secrets set ANTHROPIC_API_KEY='sk-ant-...' --app ${AI_APP}"
echo ""
echo "# Orchestrator API Key (generate a secure random key):"
echo "flyctl secrets set ORCHESTRATOR_API_KEY=\"\$(openssl rand -hex 32)\" --app ${AI_APP}"
echo ""
echo "# Optional - OpenRouter for model routing:"
echo "flyctl secrets set OPENROUTER_API_KEY='sk-or-...' --app ${AI_APP}"
echo ""

# Check if we should proceed with what we can
read -p "Do you have these values ready? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "Please gather your API keys and run this script again."
    echo "Save the commands above to a file and run them when ready."
    exit 0
fi

echo ""
echo "Great! Let's set the secrets interactively."
echo ""

# Interactive secret setting
echo "Step 1: Database URL (from ${MAIN_APP})"
echo "This should look like: postgresql://user:pass@host/db?sslmode=require"
read -p "DATABASE_URL: " db_url
if [ -n "$db_url" ]; then
    flyctl secrets set "DATABASE_URL=$db_url" --app "$AI_APP"
fi

echo ""
echo "Step 2: Redis Address (from ${MAIN_APP})"
echo "This should look like: host:6379 or xxx.upstash.io:6379"
read -p "REDIS_ADDR: " redis_addr
if [ -n "$redis_addr" ]; then
    flyctl secrets set "REDIS_ADDR=$redis_addr" --app "$AI_APP"
fi

echo ""
read -p "Does your Redis require a password? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    read -p "REDIS_PASSWORD: " redis_pass
    if [ -n "$redis_pass" ]; then
        flyctl secrets set "REDIS_PASSWORD=$redis_pass" --app "$AI_APP"
    fi
fi

echo ""
echo "Step 3: OpenAI API Key"
read -p "OPENAI_API_KEY: " openai_key
if [ -n "$openai_key" ]; then
    flyctl secrets set "OPENAI_API_KEY=$openai_key" --app "$AI_APP"
fi

echo ""
echo "Step 4: Anthropic API Key"
read -p "ANTHROPIC_API_KEY: " anthropic_key
if [ -n "$anthropic_key" ]; then
    flyctl secrets set "ANTHROPIC_API_KEY=$anthropic_key" --app "$AI_APP"
fi

echo ""
echo "Step 5: Orchestrator API Key (generating secure random key...)"
ORCH_KEY=$(openssl rand -hex 32)
echo "Generated: $ORCH_KEY"
flyctl secrets set "ORCHESTRATOR_API_KEY=$ORCH_KEY" --app "$AI_APP"
echo ""
echo "⚠️  IMPORTANT: Save this key - you'll need to set it on the orchestrator:"
echo "   flyctl secrets set AI_SERVICE_API_KEY='$ORCH_KEY' --app $MAIN_APP"

echo ""
echo "================================================"
echo "Secrets configured!"
echo "================================================"
echo ""
echo "Next steps:"
echo "  1. Deploy AI service: ./deploy.sh"
echo "  2. Update orchestrator with AI_SERVICE_URL"
echo "     flyctl secrets set AI_SERVICE_URL='http://functionfly-ai-service.internal:8081' --app $MAIN_APP"
echo "  3. Set the API key on orchestrator (use the key above)"
echo ""
