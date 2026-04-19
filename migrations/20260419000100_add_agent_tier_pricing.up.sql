-- Database-driven agent tier pricing
-- Replaces hardcoded constants from internal/plans/limits.go
CREATE TABLE IF NOT EXISTS agent_tier_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_slug VARCHAR(50) UNIQUE NOT NULL, -- 'starter', 'scale', 'pro', 'enterprise'
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Pricing (cents in base currency)
    monthly_price_cents INTEGER NOT NULL,
    annual_price_cents INTEGER, -- NULL = no annual discount
    
    -- Currency (base currency for this tier)
    base_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Multi-currency support: region-specific pricing
    region_pricing JSONB DEFAULT '{}', -- {"EUR": {"monthly": 2600, "annual": 26000}, ...}
    
    -- Limits
    max_agents INTEGER NOT NULL DEFAULT -1, -- -1 = unlimited
    included_ai_calls INTEGER NOT NULL DEFAULT 0,
    included_executions INTEGER NOT NULL DEFAULT 0,
    included_storage_gb INTEGER NOT NULL DEFAULT 0,
    
    -- Overage pricing (per 1000 units)
    overage_price_per_1000_cents INTEGER DEFAULT 0,
    
    -- Stripe integration
    stripe_price_id_monthly VARCHAR(100),
    stripe_price_id_annual VARCHAR(100),
    
    -- Feature flags
    features_included JSONB NOT NULL DEFAULT '[]',
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    sort_order INTEGER DEFAULT 0,
    
    -- For A/B testing and dynamic pricing
    pricing_variant VARCHAR(20) DEFAULT 'default', -- 'default', 'experiment_a', 'promo'
    valid_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE, -- NULL = no expiration
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed with current hardcoded pricing (can be edited in DB to change without deployment)
INSERT INTO agent_tier_pricing (
    id, tier_slug, display_name, description,
    monthly_price_cents, annual_price_cents, base_currency,
    max_agents, included_ai_calls, included_executions, included_storage_gb,
    overage_price_per_1000_cents, features_included, sort_order, is_active
) VALUES (
    gen_random_uuid(),
    'agent-starter',
    'Agent Starter',
    'Competitive entry point for small teams getting started with AI agents',
    2900, 29000, 'USD',
    5, 10000, 100000, 10,
    50,
    '["basic_agents", "webhooks", "community_support"]'::jsonb,
    1, true
) ON CONFLICT (tier_slug) DO NOTHING;

INSERT INTO agent_tier_pricing (
    id, tier_slug, display_name, description,
    monthly_price_cents, annual_price_cents, base_currency,
    max_agents, included_ai_calls, included_executions, included_storage_gb,
    overage_price_per_1000_cents, features_included, sort_order, is_active
) VALUES (
    gen_random_uuid(),
    'agent-scale',
    'Agent Scale',
    'Mid-tier with good margins for growing teams',
    14900, 149000, 'USD',
    25, 100000, 1000000, 100,
    40,
    '["advanced_agents", "workflows", "analytics", "priority_support"]'::jsonb,
    2, true
) ON CONFLICT (tier_slug) DO NOTHING;

INSERT INTO agent_tier_pricing (
    id, tier_slug, display_name, description,
    monthly_price_cents, annual_price_cents, base_currency,
    max_agents, included_ai_calls, included_executions, included_storage_gb,
    overage_price_per_1000_cents, features_included, sort_order, is_active
) VALUES (
    gen_random_uuid(),
    'agent-pro',
    'Agent Pro',
    'Professional tier for production workloads',
    39900, 399000, 'USD',
    100, 500000, 5000000, 500,
    30,
    '["all_features", "dedicated_resources", "sla", "phone_support"]'::jsonb,
    3, true
) ON CONFLICT (tier_slug) DO NOTHING;

INSERT INTO agent_tier_pricing (
    id, tier_slug, display_name, description,
    monthly_price_cents, annual_price_cents, base_currency,
    max_agents, included_ai_calls, included_executions, included_storage_gb,
    overage_price_per_1000_cents, features_included, sort_order, is_active
) VALUES (
    gen_random_uuid(),
    'agent-enterprise',
    'Agent Enterprise',
    'Custom enterprise tier with unlimited resources',
    99000, NULL, 'USD',
    -1, -1, -1, -1, -- Unlimited (-1)
    20,
    '["everything", "custom_contracts", "dedicated_infra", "white_glove"]'::jsonb,
    4, true
) ON CONFLICT (tier_slug) DO NOTHING;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_agent_tier_pricing_slug ON agent_tier_pricing(tier_slug);
CREATE INDEX IF NOT EXISTS idx_agent_tier_pricing_active ON agent_tier_pricing(is_active);
CREATE INDEX IF NOT EXISTS idx_agent_tier_pricing_variant ON agent_tier_pricing(pricing_variant);
CREATE INDEX IF NOT EXISTS idx_agent_tier_pricing_validity ON agent_tier_pricing(valid_from, valid_until);
