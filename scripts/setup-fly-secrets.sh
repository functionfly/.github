#!/bin/bash
# Script to set up Fly.io secrets for FunctionFly
# Usage: ./scripts/setup-fly-secrets.sh
# Requires: flyctl CLI installed and logged in

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up Fly.io secrets for FunctionFly...${NC}"

# Check if flyctl is installed
if ! command -v flyctl &> /dev/null; then
    echo -e "${RED}Error: flyctl is not installed${NC}"
    echo "Install with: curl -L https://fly.io/install.sh | sh"
    exit 1
fi

# Check if user is logged in
if ! flyctl auth whoami &> /dev/null; then
    echo -e "${RED}Error: Not logged in to Fly.io${NC}"
    echo "Run: flyctl auth login"
    exit 1
fi

# Load environment variables from .env if it exists
if [ -f .env ]; then
    echo -e "${YELLOW}Loading environment from .env...${NC}"
    set -a
    source .env
    set +a
fi

# Required secrets
echo -e "${YELLOW}Setting required secrets...${NC}"

# Database
if [ -n "$DATABASE_URL" ]; then
    echo "Setting DATABASE_URL..."
    echo "$DATABASE_URL" | flyctl secrets set DATABASE_URL=- --app functionfly-orchestrator
else
    echo -e "${RED}Warning: DATABASE_URL not set in environment${NC}"
fi

# Redis
if [ -n "$REDIS_URL" ] || [ -n "$REDIS_ADDR" ]; then
    REDIS_VAL="${REDIS_URL:-redis://$REDIS_ADDR}"
    echo "Setting REDIS_URL..."
    echo "$REDIS_VAL" | flyctl secrets set REDIS_URL=- --app functionfly-orchestrator
else
    echo -e "${RED}Warning: REDIS_URL/REDIS_ADDR not set in environment${NC}"
fi

# Stripe secrets
echo -e "${YELLOW}Setting Stripe secrets...${NC}"

if [ -n "$STRIPE_SECRET_KEY" ]; then
    echo "Setting STRIPE_SECRET_KEY..."
    echo "$STRIPE_SECRET_KEY" | flyctl secrets set STRIPE_SECRET_KEY=- --app functionfly-orchestrator
else
    echo -e "${RED}Warning: STRIPE_SECRET_KEY not set in environment${NC}"
fi

if [ -n "$STRIPE_PUBLISHABLE_KEY" ]; then
    echo "Setting STRIPE_PUBLISHABLE_KEY..."
    echo "$STRIPE_PUBLISHABLE_KEY" | flyctl secrets set STRIPE_PUBLISHABLE_KEY=- --app functionfly-orchestrator
else
    echo -e "${RED}Warning: STRIPE_PUBLISHABLE_KEY not set in environment${NC}"
fi

if [ -n "$STRIPE_WEBHOOK_SECRET" ]; then
    echo "Setting STRIPE_WEBHOOK_SECRET..."
    echo "$STRIPE_WEBHOOK_SECRET" | flyctl secrets set STRIPE_WEBHOOK_SECRET=- --app functionfly-orchestrator
else
    echo -e "${YELLOW}Note: STRIPE_WEBHOOK_SECRET not set (optional for local testing)${NC}"
fi

# Stripe Price IDs for State Fabric Addons
echo -e "${YELLOW}Setting Stripe Price IDs...${NC}"

[ -n "$STRIPE_PRICE_SF_ADDON_HOT_CACHE" ] && \
    echo "$STRIPE_PRICE_SF_ADDON_HOT_CACHE" | flyctl secrets set STRIPE_PRICE_SF_ADDON_HOT_CACHE=- --app functionfly-orchestrator

[ -n "$STRIPE_PRICE_SF_ADDON_SECURITY" ] && \
    echo "$STRIPE_PRICE_SF_ADDON_SECURITY" | flyctl secrets set STRIPE_PRICE_SF_ADDON_SECURITY=- --app functionfly-orchestrator

[ -n "$STRIPE_PRICE_SF_ADDON_AI_MEMORY" ] && \
    echo "$STRIPE_PRICE_SF_ADDON_AI_MEMORY" | flyctl secrets set STRIPE_PRICE_SF_ADDON_AI_MEMORY=- --app functionfly-orchestrator

[ -n "$STRIPE_PRICE_SF_ADDON_INSIGHTS" ] && \
    echo "$STRIPE_PRICE_SF_ADDON_INSIGHTS" | flyctl secrets set STRIPE_PRICE_SF_ADDON_INSIGHTS=- --app functionfly-orchestrator

# JWT secrets
if [ -n "$JWT_SECRET" ]; then
    echo "Setting JWT_SECRET..."
    echo "$JWT_SECRET" | flyctl secrets set JWT_SECRET=- --app functionfly-orchestrator
else
    echo -e "${RED}Warning: JWT_SECRET not set - generating new secret${NC}"
    NEW_JWT_SECRET=$(openssl rand -base64 32)
    echo "$NEW_JWT_SECRET" | flyctl secrets set JWT_SECRET=- --app functionfly-orchestrator
fi

if [ -n "$JWT_REFRESH_SECRET" ]; then
    echo "Setting JWT_REFRESH_SECRET..."
    echo "$JWT_REFRESH_SECRET" | flyctl secrets set JWT_REFRESH_SECRET=- --app functionfly-orchestrator
else
    echo -e "${RED}Warning: JWT_REFRESH_SECRET not set - using JWT_SECRET${NC}"
fi

# Storage secrets (GCS)
echo -e "${YELLOW}Setting GCS secrets...${NC}"

if [ -n "$GCS_CREDENTIALS" ]; then
    echo "Setting GCS_CREDENTIALS..."
    echo "$GCS_CREDENTIALS" | flyctl secrets set GCS_CREDENTIALS=- --app functionfly-orchestrator
else
    echo -e "${YELLOW}Note: GCS_CREDENTIALS not set (required for GCS storage)${NC}"
fi

# Email (Resend)
if [ -n "$RESEND_API_KEY" ]; then
    echo "Setting RESEND_API_KEY..."
    echo "$RESEND_API_KEY" | flyctl secrets set RESEND_API_KEY=- --app functionfly-orchestrator
else
    echo -e "${YELLOW}Note: RESEND_API_KEY not set (email functionality disabled)${NC}"
fi

# AI Service
if [ -n "$OPENAI_API_KEY" ]; then
    echo "Setting OPENAI_API_KEY..."
    echo "$OPENAI_API_KEY" | flyctl secrets set OPENAI_API_KEY=- --app functionfly-orchestrator
else
    echo -e "${YELLOW}Note: OPENAI_API_KEY not set (AI features disabled)${NC}"
fi

# Fern API Keys
if [ -n "$FERN_API_KEY" ]; then
    echo "Setting FERN_API_KEY..."
    echo "$FERN_API_KEY" | flyctl secrets set FERN_API_KEY=- --app functionfly-orchestrator
fi

echo -e "${GREEN}Done! Secrets have been set.${NC}"
echo ""
echo "To verify secrets are set, run:"
echo "  flyctl secrets list --app functionfly-orchestrator"
echo ""
echo "To deploy with the new secrets:"
echo "  flyctl deploy --app functionfly-orchestrator"
