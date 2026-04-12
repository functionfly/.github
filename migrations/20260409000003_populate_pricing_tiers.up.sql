-- Migration: 20260409000003_populate_pricing_tiers
-- Description: Populate pricing_tiers table with correct Stripe price IDs
-- Problem: The pricing_tiers table is empty, causing checkout failures
-- Solution: Insert/update pricing tiers with valid Stripe price IDs

BEGIN;

-- First, update any existing pricing tiers that have invalid stripe_price_id values
-- This fixes the issue where product IDs (prod_*) were incorrectly stored
UPDATE pricing_tiers
SET
    stripe_price_id = NULL,
    updated_at = NOW()
WHERE stripe_price_id LIKE 'prod_%';

-- Insert or update the Starter tier
-- Use a DO block to handle the upsert since we can't rely on a unique constraint
DO $$
BEGIN
    -- Check if a tier with name 'Starter' exists
    IF EXISTS (SELECT 1 FROM pricing_tiers WHERE name = 'Starter') THEN
        UPDATE pricing_tiers
        SET
            description = 'For side projects and MVPs. Get started with core features.',
            price_cents = 2900,
            currency = 'USD',
            features = '{"functions": 5, "providers": 3, "requests": 1000000, "agents": 1, "storage_gb": 1, "state_fabrics": 1}'::jsonb,
            is_active = true,
            tier_type = 'subscription',
            stripe_price_id = 'price_1TF3rMKxe78JyppiwBOdgrl8', -- From VITE_STRIPE_PRICE_STARTER
            trial_days = 14,
            max_agents = 1,
            max_functions = 5,
            max_executions_per_month = 1000000,
            updated_at = NOW()
        WHERE name = 'Starter';
    ELSE
        INSERT INTO pricing_tiers (
            id, name, description, price_cents, currency, features, is_active,
            tier_type, stripe_price_id, trial_days, max_agents, max_functions, max_executions_per_month,
            created_at, updated_at
        ) VALUES (
            gen_random_uuid(), 'Starter', 'For side projects and MVPs. Get started with core features.',
            2900, 'USD',
            '{"functions": 5, "providers": 3, "requests": 1000000, "agents": 1, "storage_gb": 1, "state_fabrics": 1}'::jsonb,
            true, 'subscription', 'price_1TF3rMKxe78JyppiwBOdgrl8', 14, 1, 5, 1000000, NOW(), NOW()
        );
    END IF;
END $$;

-- Insert or update the Professional tier
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pricing_tiers WHERE name = 'Professional') THEN
        UPDATE pricing_tiers
        SET
            description = 'For growing teams and production workloads.',
            price_cents = 9900,
            currency = 'USD',
            features = '{"functions": 25, "providers": 5, "requests": 5000000, "agents": 5, "storage_gb": 10, "state_fabrics": 3}'::jsonb,
            is_active = true,
            tier_type = 'subscription',
            stripe_price_id = 'price_1TF3v8Kxe78JyppidfZgdu7Z', -- From VITE_STRIPE_PRICE_PROFESSIONAL
            trial_days = 14,
            max_agents = 5,
            max_functions = 25,
            max_executions_per_month = 5000000,
            updated_at = NOW()
        WHERE name = 'Professional';
    ELSE
        INSERT INTO pricing_tiers (
            id, name, description, price_cents, currency, features, is_active,
            tier_type, stripe_price_id, trial_days, max_agents, max_functions, max_executions_per_month,
            created_at, updated_at
        ) VALUES (
            gen_random_uuid(), 'Professional', 'For growing teams and production workloads.',
            9900, 'USD',
            '{"functions": 25, "providers": 5, "requests": 5000000, "agents": 5, "storage_gb": 10, "state_fabrics": 3}'::jsonb,
            true, 'subscription', 'price_1TF3v8Kxe78JyppidfZgdu7Z', 14, 5, 25, 5000000, NOW(), NOW()
        );
    END IF;
END $$;

-- Insert or update the Free tier
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pricing_tiers WHERE name = 'Free') THEN
        UPDATE pricing_tiers
        SET
            description = 'Perfect for getting started with FunctionFly.',
            price_cents = 0,
            currency = 'USD',
            features = '{"functions": 1, "providers": 2, "requests": 100000, "agents": 0, "storage_gb": 0, "state_fabrics": 0}'::jsonb,
            is_active = true,
            tier_type = 'subscription',
            stripe_price_id = NULL, -- Free tier has no stripe price ID
            trial_days = 0,
            max_agents = 0,
            max_functions = 1,
            max_executions_per_month = 100000,
            updated_at = NOW()
        WHERE name = 'Free';
    ELSE
        INSERT INTO pricing_tiers (
            id, name, description, price_cents, currency, features, is_active,
            tier_type, stripe_price_id, trial_days, max_agents, max_functions, max_executions_per_month,
            created_at, updated_at
        ) VALUES (
            gen_random_uuid(), 'Free', 'Perfect for getting started with FunctionFly.',
            0, 'USD',
            '{"functions": 1, "providers": 2, "requests": 100000, "agents": 0, "storage_gb": 0, "state_fabrics": 0}'::jsonb,
            true, 'subscription', NULL, 0, 0, 1, 100000, NOW(), NOW()
        );
    END IF;
END $$;

COMMIT;
