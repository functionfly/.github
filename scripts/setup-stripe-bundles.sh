#!/bin/bash
# Stripe Bundle Setup Script
# Creates products and prices for Backend-in-a-Box bundles

set -e

# Check for Stripe secret key
if [ -z "$STRIPE_SECRET_KEY" ]; then
    echo "Error: STRIPE_SECRET_KEY environment variable not set"
    exit 1
fi

echo "Creating Stripe products and prices for Backend-in-a-Box bundles..."

# SaaS Starter Pack - $29/month
echo "Creating SaaS Starter Pack..."
SAAS_PRODUCT=$(curl -s -X POST https://api.stripe.com/v1/products \
  -u "$STRIPE_SECRET_KEY:" \
  -d name="SaaS Starter Pack" \
  -d description="Pre-configured SaaS backend: Auth, Payments, User DB, Email, Analytics. One click → full backend." \
  -d "metadata[bundle_slug]=saas-starter")

SAAS_PRODUCT_ID=$(echo $SAAS_PRODUCT | grep -o '"id": "prod_[^"]*' | head -1 | cut -d'"' -f4)
echo "  Product ID: $SAAS_PRODUCT_ID"

SAAS_PRICE=$(curl -s -X POST https://api.stripe.com/v1/prices \
  -u "$STRIPE_SECRET_KEY:" \
  -d product="$SAAS_PRODUCT_ID" \
  -d unit_amount=2900 \
  -d currency=usd \
  -d "recurring[interval]=month" \
  -d "metadata[bundle_slug]=saas-starter")

SAAS_PRICE_ID=$(echo $SAAS_PRICE | grep -o '"id": "price_[^"]*' | head -1 | cut -d'"' -f4)
echo "  Price ID: $SAAS_PRICE_ID"

# Marketplace Pack - $49/month
echo "Creating Marketplace Pack..."
MP_PRODUCT=$(curl -s -X POST https://api.stripe.com/v1/products \
  -u "$STRIPE_SECRET_KEY:" \
  -d name="Marketplace Pack" \
  -d description="Complete marketplace backend: Listings, Payments, Messaging, Notifications." \
  -d "metadata[bundle_slug]=marketplace")

MP_PRODUCT_ID=$(echo $MP_PRODUCT | grep -o '"id": "prod_[^"]*' | head -1 | cut -d'"' -f4)
echo "  Product ID: $MP_PRODUCT_ID"

MP_PRICE=$(curl -s -X POST https://api.stripe.com/v1/prices \
  -u "$STRIPE_SECRET_KEY:" \
  -d product="$MP_PRODUCT_ID" \
  -d unit_amount=4900 \
  -d currency=usd \
  -d "recurring[interval]=month" \
  -d "metadata[bundle_slug]=marketplace")

MP_PRICE_ID=$(echo $MP_PRICE | grep -o '"id": "price_[^"]*' | head -1 | cut -d'"' -f4)
echo "  Price ID: $MP_PRICE_ID"

# AI App Pack - $39/month
echo "Creating AI App Pack..."
AI_PRODUCT=$(curl -s -X POST https://api.stripe.com/v1/products \
  -u "$STRIPE_SECRET_KEY:" \
  -d name="AI App Pack" \
  -d description="AI infrastructure: Vector DB, Embeddings, Chat workflows, Memory system." \
  -d "metadata[bundle_slug]=ai-app")

AI_PRODUCT_ID=$(echo $AI_PRODUCT | grep -o '"id": "prod_[^"]*' | head -1 | cut -d'"' -f4)
echo "  Product ID: $AI_PRODUCT_ID"

AI_PRICE=$(curl -s -X POST https://api.stripe.com/v1/prices \
  -u "$STRIPE_SECRET_KEY:" \
  -d product="$AI_PRODUCT_ID" \
  -d unit_amount=3900 \
  -d currency=usd \
  -d "recurring[interval]=month" \
  -d "metadata[bundle_slug]=ai-app")

AI_PRICE_ID=$(echo $AI_PRICE | grep -o '"id": "price_[^"]*' | head -1 | cut -d'"' -f4)
echo "  Price ID: $AI_PRICE_ID"

echo ""
echo "=========================================="
echo "Add these to your .env file:"
echo ""
echo "STRIPE_PRICE_SAAS_STARTER=$SAAS_PRICE_ID"
echo "STRIPE_PRICE_MARKETPLACE=$MP_PRICE_ID"
echo "STRIPE_PRICE_AI_APP=$AI_PRICE_ID"
echo ""
echo "Or update the database directly:"
echo ""
echo "UPDATE pricing_bundles SET stripe_price_id = '$SAAS_PRICE_ID' WHERE slug = 'saas-starter';"
echo "UPDATE pricing_bundles SET stripe_price_id = '$MP_PRICE_ID' WHERE slug = 'marketplace';"
echo "UPDATE pricing_bundles SET stripe_price_id = '$AI_PRICE_ID' WHERE slug = 'ai-app';"
echo "=========================================="
