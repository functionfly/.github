-- Migration: 20260426185721_unified_plans_2026
-- Description: Merge agent plans into main plans with 2026 pricing
-- This migration aliases old agent plan tiers to main plan equivalents and
-- updates pricing tiers to reflect the new unified plan structure.

BEGIN;

-- =============================================================================
-- Step 1: Create plan alias mapping table for backward compatibility
-- =============================================================================

CREATE TABLE IF NOT EXISTS plan_aliases (
    legacy_plan_name VARCHAR(50) PRIMARY KEY,
    unified_plan_name VARCHAR(50) NOT NULL,
    is_deprecated BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert plan alias mappings
-- Old agent tiers now map to main plan equivalents
INSERT INTO plan_aliases (legacy_plan_name, unified_plan_name, is_deprecated) VALUES
    ('agent_starter', 'starter', true),
    ('agent_scale', 'professional', true),
    ('agent_pro', 'enterprise', true),
    ('agent_enterprise', 'agent_enterprise', false)  -- Still active as max tier
ON CONFLICT (legacy_plan_name) DO NOTHING;

-- =============================================================================
-- Step 2: Update pricing_tiers table with new 2026 pricing
-- =============================================================================

-- Update Starter plan price ($29 -> $24)
UPDATE pricing_tiers
SET
    price_cents = 2400,
    annual_price_cents = 24000,
    description = 'For side projects and MVPs with bundled AI agents (100K calls/mo, 10 concurrency)'
WHERE name = 'starter';

-- Update Professional plan price ($99 -> $79)
UPDATE pricing_tiers
SET
    price_cents = 7900,
    annual_price_cents = 79000,
    description = 'For growing businesses with AI agents (1M calls/mo, 100 concurrency)'
WHERE name = 'professional';

-- Update Enterprise plan ($499 -> $199, now has defined pricing)
UPDATE pricing_tiers
SET
    price_cents = 19900,
    annual_price_cents = 199000,
    description = 'For large-scale applications with AI agents (10M calls/mo, 500 concurrency)'
WHERE name = 'enterprise'
AND price_cents = 49900;  -- Only update if it's still the old price

-- Add Agent Enterprise tier if it doesn't exist
INSERT INTO pricing_tiers (name, display_name, price_cents, annual_price_cents, description, billing_cycle, is_active, tier_type)
SELECT 'agent_enterprise', 'Agent Enterprise', 49900, 499000, 'Unlimited AI agents for enterprise scale', 'monthly', true, 'subscription'
WHERE NOT EXISTS (SELECT 1 FROM pricing_tiers WHERE name = 'agent_enterprise');

-- =============================================================================
-- Step 3: Update limits in pricing_tiers table
-- =============================================================================

-- Update starter limits to include AI agent capabilities
UPDATE pricing_tiers
SET
    max_agents = 10,
    features = '{"ai_calls_per_month": 100000, "agent_concurrency": 10, "agent_calls_per_minute": 100, "state_writes_per_hour": 1000}'
WHERE name = 'starter';

-- Update professional limits
UPDATE pricing_tiers
SET
    max_agents = 100,
    features = '{"ai_calls_per_month": 1000000, "agent_concurrency": 100, "agent_calls_per_minute": 500, "state_writes_per_hour": 10000}'
WHERE name = 'professional';

-- Update enterprise limits
UPDATE pricing_tiers
SET
    max_agents = 500,
    features = '{"ai_calls_per_month": 10000000, "agent_concurrency": 500, "agent_calls_per_minute": 2000, "state_writes_per_hour": 50000}'
WHERE name = 'enterprise';

-- Update agent_enterprise limits (unlimited)
UPDATE pricing_tiers
SET
    max_agents = -1,
    features = '{"ai_calls_per_month": -1, "agent_concurrency": -1, "agent_calls_per_minute": -1, "state_writes_per_hour": -1, "unlimited": true}'
WHERE name = 'agent_enterprise';

-- =============================================================================
-- Step 4: Create overage pricing entries
-- =============================================================================

CREATE TABLE IF NOT EXISTS overage_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_name VARCHAR(50) NOT NULL,
    rate_type VARCHAR(20) NOT NULL DEFAULT 'ai_calls',  -- 'ai_calls', 'requests', 'storage'
    cents_per_thousand INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(plan_name, rate_type)
);

-- Insert overage rates for AI calls (cents per 1000)
INSERT INTO overage_rates (plan_name, rate_type, cents_per_thousand) VALUES
    ('starter', 'ai_calls', 15),      -- $0.15 per 1K calls
    ('professional', 'ai_calls', 8),  -- $0.08 per 1K calls
    ('enterprise', 'ai_calls', 4),     -- $0.04 per 1K calls
    ('agent_enterprise', 'ai_calls', 0)  -- No overage (unlimited)
ON CONFLICT (plan_name, rate_type) DO UPDATE SET cents_per_thousand = EXCLUDED.cents_per_thousand;

-- =============================================================================
-- Step 5: Update agent_tier_pricing to reflect new pricing
-- =============================================================================

UPDATE agent_tier_pricing
SET
    monthly_price_cents = CASE tier_slug
        WHEN 'agent-starter' THEN 2400
        WHEN 'agent-scale' THEN 7900
        WHEN 'agent-pro' THEN 19900
        WHEN 'agent-enterprise' THEN 49900
        ELSE monthly_price_cents
    END,
    display_name = CASE tier_slug
        WHEN 'agent-starter' THEN 'Agent Starter (Starter Plan)'
        WHEN 'agent-scale' THEN 'Agent Scale (Professional Plan)'
        WHEN 'agent-pro' THEN 'Agent Pro (Enterprise Plan)'
        ELSE display_name
    END
WHERE tier_slug IN ('agent-starter', 'agent-scale', 'agent-pro', 'agent-enterprise');

COMMIT;

-- =============================================================================
-- Verification queries (run separately to check)
-- =============================================================================

-- SELECT * FROM plan_aliases;
-- SELECT name, price_cents, annual_price_cents, max_agents FROM pricing_tiers WHERE name IN ('starter', 'professional', 'enterprise', 'agent_enterprise');
-- SELECT * FROM overage_rates WHERE rate_type = 'ai_calls';