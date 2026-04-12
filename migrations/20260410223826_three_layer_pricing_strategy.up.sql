-- Migration: 20260410223826_three_layer_pricing_strategy
-- Description: Implement 3-layer pricing (Free, Usage-Based, Pro/Team) with generous free tier and metered billing

-- ============================================
-- LAYER 1: FREE TIER - Generous limits for growth
-- ============================================

-- Update Free tier to be ridiculously generous
UPDATE pricing_tiers SET
    max_functions = 10,           -- Increased from 5
    max_executions_per_month = 100000,  -- 100k executions/month
    trial_days = 0,
    features = jsonb_build_object(
        'requests', 100000,
        'included_compute_ms', 3600000,  -- 1 hour compute time
        'included_compute_hours', 1.0,
        'functions', 10,
        'workflows', 5,
        'logs_retention_days', 7,
        'community_functions_access', true,
        'basic_workflows', true,
        'ai_calls_included', 1000,
        'storage_gb', 1,
        'vector_storage_entries', 1000,
        'support', 'community'
    )
WHERE name = 'Free';

-- ============================================
-- LAYER 2 & 3: USAGE-BASED METERING TABLES
-- ============================================

-- Extended usage events table for granular metering
CREATE TABLE IF NOT EXISTS usage_events_v2 (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    event_type VARCHAR(50) NOT NULL, -- 'function_execution', 'ai_call', 'state_read', 'state_write', 'vector_query', 'workflow_run'
    quantity INTEGER NOT NULL DEFAULT 1,
    
    -- For AI calls: track model and tokens
    ai_model VARCHAR(50),           -- e.g., 'gpt-4', 'claude-3-opus'
    ai_input_tokens INTEGER,
    ai_output_tokens INTEGER,
    ai_cost_usd DECIMAL(12, 6),     -- Actual cost from provider
    
    -- For state operations
    state_operation VARCHAR(20),    -- 'read', 'write', 'query', 'vector_search'
    storage_bytes INTEGER,          -- Data size for storage operations
    
    -- For workflow runs
    workflow_id UUID,               -- Reference to workflow
    workflow_complexity VARCHAR(20), -- 'simple', 'standard', 'complex'
    
    -- Cost tracking for usage-based billing
    unit_price_cents INTEGER,       -- Price per unit at time of event
    calculated_cost_cents INTEGER,  -- quantity * unit_price_cents
    
    metadata JSONB,                 -- Additional event data
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Usage rollups v2 with cost tracking
CREATE TABLE IF NOT EXISTS usage_rollups_v2 (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    event_type VARCHAR(50) NOT NULL,
    period_date DATE NOT NULL,
    
    -- Aggregated metrics
    total_quantity INTEGER NOT NULL DEFAULT 0,
    total_cost_cents INTEGER DEFAULT 0,  -- For usage-based billing
    
    -- AI-specific rollups
    ai_total_tokens INTEGER DEFAULT 0,
    ai_total_cost_usd DECIMAL(12, 6) DEFAULT 0,
    
    -- State-specific rollups
    state_reads INTEGER DEFAULT 0,
    state_writes INTEGER DEFAULT 0,
    storage_gb DECIMAL(10, 4) DEFAULT 0,
    
    -- Workflow rollups
    workflow_runs INTEGER DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, event_type, period_date)
);

-- ============================================
-- USAGE-BASED PRICING CONFIGURATION
-- ============================================

-- Pay-as-you-go pricing config (layer 2)
CREATE TABLE IF NOT EXISTS usage_pricing_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL UNIQUE, -- 'function_execution', 'ai_call', 'state_read', etc.
    
    -- Tiered pricing: JSON array of {min_quantity, max_quantity, unit_price_cents}
    -- Example: [{"min": 0, "max": 1000000, "price": 20}, {"min": 1000001, "max": null, "price": 15}]
    tiered_pricing JSONB NOT NULL DEFAULT '[]',
    
    -- AI markup settings
    ai_markup_percent INTEGER DEFAULT 25,  -- 25% markup on AI costs
    ai_minimum_charge_cents INTEGER DEFAULT 1,  -- Minimum 1 cent per AI call
    
    -- State pricing
    state_read_price_cents INTEGER DEFAULT 0,   -- Often free
    state_write_price_cents INTEGER DEFAULT 1,  -- 1 cent per 1000 writes
    storage_gb_price_cents INTEGER DEFAULT 50,  -- $0.50 per GB/month
    vector_query_price_cents INTEGER DEFAULT 5, -- 5 cents per 1000 queries
    
    -- Workflow pricing
    workflow_simple_price_cents INTEGER DEFAULT 1,   -- 1 cent per run
    workflow_standard_price_cents INTEGER DEFAULT 2, -- 2 cents per run
    workflow_complex_price_cents INTEGER DEFAULT 5,  -- 5 cents per run
    
    is_active BOOLEAN DEFAULT true,
    effective_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    effective_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default usage pricing configuration
INSERT INTO usage_pricing_config (event_type, tiered_pricing, ai_markup_percent) VALUES
    ('function_execution', '[{"min": 0, "max": 100000, "price": 0}, {"min": 100001, "max": 1000000, "price": 20}, {"min": 1000001, "max": 10000000, "price": 15}, {"min": 10000001, "max": null, "price": 10}]'::jsonb, 0),
    ('ai_call', '[]'::jsonb, 25),  -- Pass-through + 25%
    ('state_read', '[{"min": 0, "max": null, "price": 0}]'::jsonb, 0),
    ('state_write', '[{"min": 0, "max": 10000, "price": 0}, {"min": 10001, "max": null, "price": 1}]'::jsonb, 0),
    ('vector_search', '[{"min": 0, "max": 1000, "price": 0}, {"min": 1001, "max": null, "price": 5}]'::jsonb, 0),
    ('workflow_run', '[{"min": 0, "max": 1000, "price": 0}, {"min": 1001, "max": null, "price": 2}]'::jsonb, 0)
ON CONFLICT (event_type) DO NOTHING;

-- ============================================
-- PRO/TEAM TIERS (Layer 3 - High Margin)
-- ============================================

-- Create Pro tier if not exists
INSERT INTO pricing_tiers (
    id, name, tier_type, stripe_price_id, trial_days,
    max_agents, max_functions, max_executions_per_month,
    price_cents, currency,
    features
) VALUES (
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    'Pro',
    'subscription',
    NULL,  -- Set this to your Stripe price ID
    14,    -- 14-day trial
    10,    -- 10 agents
    100,   -- 100 functions
    10000000,  -- 10M executions (soft limit, overage billed)
    4900,  -- $49/month
    'USD',
    jsonb_build_object(
        'requests', 10000000,
        'included_compute_ms', 360000000,  -- 100 hours
        'included_compute_hours', 100.0,
        'functions', 100,
        'workflows', 50,
        'logs_retention_days', 30,
        'community_functions_access', true,
        'advanced_workflows', true,
        'ai_calls_included', 10000,
        'ai_overage_rate', 'pass_through_plus_20',  -- 20% markup instead of 25%
        'storage_gb', 10,
        'vector_storage_entries', 50000,
        'priority_support', true,
        'custom_domains', true,
        'webhook_customization', true,
        'team_members', 5
    )
) ON CONFLICT (id) DO UPDATE SET
    max_functions = EXCLUDED.max_functions,
    max_executions_per_month = EXCLUDED.max_executions_per_month,
    price_cents = EXCLUDED.price_cents,
    features = EXCLUDED.features;

-- Create Team tier if not exists
INSERT INTO pricing_tiers (
    id, name, tier_type, stripe_price_id, trial_days,
    max_agents, max_functions, max_executions_per_month,
    price_cents, currency,
    features
) VALUES (
    'b2c3d4e5-f6a7-8901-bcde-f23456789012',
    'Team',
    'subscription',
    NULL,  -- Set this to your Stripe price ID
    14,
    50,    -- 50 agents
    500,   -- 500 functions
    50000000,  -- 50M executions
    19900, -- $199/month
    'USD',
    jsonb_build_object(
        'requests', 50000000,
        'included_compute_ms', 1800000000,  -- 500 hours
        'included_compute_hours', 500.0,
        'functions', 500,
        'workflows', 200,
        'logs_retention_days', 90,
        'community_functions_access', true,
        'advanced_workflows', true,
        'ai_calls_included', 50000,
        'ai_overage_rate', 'pass_through_plus_15',  -- 15% markup
        'storage_gb', 100,
        'vector_storage_entries', 500000,
        'priority_support', true,
        'custom_domains', true,
        'webhook_customization', true,
        'team_members', 20,
        'sso', true,
        'audit_logs', true,
        'dedicated_resources', false
    )
) ON CONFLICT (id) DO UPDATE SET
    max_functions = EXCLUDED.max_functions,
    max_executions_per_month = EXCLUDED.max_executions_per_month,
    price_cents = EXCLUDED.price_cents,
    features = EXCLUDED.features;

-- ============================================
-- BILLING INVOICE ITEMS FOR USAGE-BASED CHARGES
-- ============================================

-- Track pending charges for usage-based billing
CREATE TABLE IF NOT EXISTS pending_usage_charges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    
    charge_type VARCHAR(50) NOT NULL, -- 'execution_overage', 'ai_calls', 'state_usage', 'workflow_runs'
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    quantity INTEGER NOT NULL,
    unit_price_cents INTEGER NOT NULL,
    total_cents INTEGER NOT NULL,
    
    -- Breakdown for transparency
    breakdown JSONB,  -- Detailed breakdown by sub-type
    
    is_invoiced BOOLEAN DEFAULT false,
    invoice_id UUID REFERENCES invoices(id),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(tenant_id, charge_type, period_start)
);

-- ============================================
-- INDEXES FOR PERFORMANCE
-- ============================================

CREATE INDEX IF NOT EXISTS idx_usage_events_v2_tenant_id ON usage_events_v2(tenant_id);
CREATE INDEX IF NOT EXISTS idx_usage_events_v2_event_type ON usage_events_v2(event_type);
CREATE INDEX IF NOT EXISTS idx_usage_events_v2_timestamp ON usage_events_v2(timestamp);
CREATE INDEX IF NOT EXISTS idx_usage_events_v2_tenant_event_time ON usage_events_v2(tenant_id, event_type, timestamp);

CREATE INDEX IF NOT EXISTS idx_usage_rollups_v2_tenant_id ON usage_rollups_v2(tenant_id);
CREATE INDEX IF NOT EXISTS idx_usage_rollups_v2_period ON usage_rollups_v2(period_date);
CREATE INDEX IF NOT EXISTS idx_usage_rollups_v2_tenant_period ON usage_rollups_v2(tenant_id, period_date);

CREATE INDEX IF NOT EXISTS idx_pending_charges_tenant ON pending_usage_charges(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pending_charges_period ON pending_usage_charges(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_pending_charges_uninvoiced ON pending_usage_charges(tenant_id, is_invoiced) WHERE is_invoiced = false;

-- ============================================
-- BILLING CALCULATION FUNCTION
-- ============================================

-- Function to calculate usage-based charges for a tenant
CREATE OR REPLACE FUNCTION calculate_usage_charges(
    p_tenant_id UUID,
    p_period_start DATE,
    p_period_end DATE
) RETURNS TABLE (
    charge_type VARCHAR,
    quantity INTEGER,
    unit_price_cents INTEGER,
    total_cents INTEGER,
    breakdown JSONB
) AS $$
BEGIN
    -- Function execution overage (above free tier 100k)
    RETURN QUERY
    SELECT 
        'execution_overage'::VARCHAR,
        GREATEST(0, COALESCE(SUM(ur.total_quantity), 0) - 100000)::INTEGER,
        20::INTEGER,  -- $0.20 per million = 0.00002 cents per execution
        (GREATEST(0, COALESCE(SUM(ur.total_quantity), 0) - 100000)::INTEGER * 20 / 1000000)::INTEGER,
        jsonb_build_object(
            'total_executions', COALESCE(SUM(ur.total_quantity), 0),
            'free_included', 100000,
            'billable', GREATEST(0, COALESCE(SUM(ur.total_quantity), 0) - 100000)
        )
    FROM usage_rollups_v2 ur
    WHERE ur.tenant_id = p_tenant_id
      AND ur.event_type = 'function_execution'
      AND ur.period_date BETWEEN p_period_start AND p_period_end
    HAVING GREATEST(0, COALESCE(SUM(ur.total_quantity), 0) - 100000) > 0;
    
    -- AI calls with markup (if free tier 1000 exceeded)
    RETURN QUERY
    SELECT 
        'ai_calls'::VARCHAR,
        GREATEST(0, COALESCE(SUM(ur.total_quantity), 0) - 1000)::INTEGER,
        0::INTEGER,  -- Variable pricing
        (COALESCE(SUM(ur.total_cost_cents), 0) * 1.25)::INTEGER,  -- 25% markup
        jsonb_build_object(
            'total_calls', COALESCE(SUM(ur.total_quantity), 0),
            'free_included', 1000,
            'base_cost_cents', COALESCE(SUM(ur.total_cost_cents), 0),
            'markup_percent', 25
        )
    FROM usage_rollups_v2 ur
    WHERE ur.tenant_id = p_tenant_id
      AND ur.event_type = 'ai_call'
      AND ur.period_date BETWEEN p_period_start AND p_period_end
    HAVING GREATEST(0, COALESCE(SUM(ur.total_quantity), 0) - 1000) > 0;
    
    -- State storage (if above 1GB free tier)
    RETURN QUERY
    SELECT 
        'state_storage'::VARCHAR,
        GREATEST(0, (COALESCE(MAX(ur.storage_gb), 0) * 10000)::INTEGER - 10000)::INTEGER,  -- Convert to 0.0001 GB units
        50::INTEGER,  -- $0.50 per GB = 50 cents per 1GB
        (GREATEST(0, COALESCE(MAX(ur.storage_gb), 0)::INTEGER - 1) * 50)::INTEGER,
        jsonb_build_object(
            'peak_storage_gb', COALESCE(MAX(ur.storage_gb), 0),
            'free_included_gb', 1,
            'billable_gb', GREATEST(0, COALESCE(MAX(ur.storage_gb), 0)::INTEGER - 1)
        )
    FROM usage_rollups_v2 ur
    WHERE ur.tenant_id = p_tenant_id
      AND ur.event_type = 'state_write'
      AND ur.period_date BETWEEN p_period_start AND p_period_end
    HAVING GREATEST(0, COALESCE(MAX(ur.storage_gb), 0)::INTEGER - 1) > 0;
    
    -- Workflow runs (if above 1000 free tier)
    RETURN QUERY
    SELECT 
        'workflow_runs'::VARCHAR,
        GREATEST(0, COALESCE(SUM(ur.workflow_runs), 0) - 1000)::INTEGER,
        2::INTEGER,  -- $0.02 per 100 = 2 cents per 100
        (GREATEST(0, COALESCE(SUM(ur.workflow_runs), 0) - 1000)::INTEGER * 2 / 100)::INTEGER,
        jsonb_build_object(
            'total_runs', COALESCE(SUM(ur.workflow_runs), 0),
            'free_included', 1000,
            'billable', GREATEST(0, COALESCE(SUM(ur.workflow_runs), 0) - 1000)
        )
    FROM usage_rollups_v2 ur
    WHERE ur.tenant_id = p_tenant_id
      AND ur.event_type = 'workflow_run'
      AND ur.period_date BETWEEN p_period_start AND p_period_end
    HAVING GREATEST(0, COALESCE(SUM(ur.workflow_runs), 0) - 1000) > 0;
END;
$$ LANGUAGE plpgsql;
