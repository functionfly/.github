#!/bin/bash
# update-bundle-price-ids.sh
# Updates Stripe Price IDs for Backend-in-a-Box bundles from environment variables
# Run this after migrations to set actual Stripe Price IDs

set -e

echo "Updating Bundle Stripe Price IDs..."

# Check if required environment variables are set
if [ -z "$DATABASE_URL" ] && [ -z "$DB_HOST" ]; then
    echo "Error: DATABASE_URL or DB_HOST must be set"
    exit 1
fi

# Build connection string if not using DATABASE_URL
if [ -z "$DATABASE_URL" ]; then
    export PGHOST="${DB_HOST:-localhost}"
    export PGPORT="${DB_PORT:-5432}"
    export PGUSER="${DB_USER:-postgres}"
    export PGPASSWORD="${DB_PASSWORD:-}"
    export PGDATABASE="${DB_NAME:-functionfly}"
fi

# Function to update price ID
update_price_id() {
    local slug=$1
    local price_id=$2
    
    if [ -n "$price_id" ]; then
        echo "Updating $slug bundle with price ID: $price_id"
        if [ -n "$DATABASE_URL" ]; then
            psql "$DATABASE_URL" -c "UPDATE pricing_bundles SET stripe_price_id = '$price_id' WHERE slug = '$slug' AND (stripe_price_id IS NULL OR stripe_price_id = '');"
        else
            psql -c "UPDATE pricing_bundles SET stripe_price_id = '$price_id' WHERE slug = '$slug' AND (stripe_price_id IS NULL OR stripe_price_id = '');"
        fi
    else
        echo "Warning: STRIPE_PRICE_${slug} not set, skipping $slug bundle"
    fi
}

# Update each bundle
update_price_id "saas-starter" "$STRIPE_PRICE_SAAS_STARTER"
update_price_id "marketplace" "$STRIPE_PRICE_MARKETPLACE"
update_price_id "ai-app" "$STRIPE_PRICE_AI_APP"

echo "Bundle Price ID update complete!"
