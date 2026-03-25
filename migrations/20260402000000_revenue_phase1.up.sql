-- Migration: 20260402000000_revenue_phase1
-- Description: Revenue System Phase 1 - Trust Layer Monetization
-- Created: 2026-04-02
-- Part of: Moat Competitive Analysis Phase 1 - Revenue Foundation
-- Features: Pricing Tiers (Hobby/Pro/Scale/Enterprise), Verification Fees, Marketplace Cut, Agent Subscriptions

BEGIN;

-- ============================================
-- Pricing Tiers (expanded from basic)
-- Standard tiers: Hobby (free), Pro ($49), Scale ($299), Enterprise (custom)
-- ============================================

ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS tier_type VARCHAR(20) DEFAULT 'subscription';
ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS stripe_price_id VARCHAR(100);
ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS trial_days INTEGER DEFAULT 0;
ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS max_agents INTEGER DEFAULT 1;
ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS max_functions INTEGER DEFAULT 10;
ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS max_executions_per_month INTEGER DEFAULT 1000;

-- Update existing pricing tiers with new structure
UPDATE pricing_tiers SET tier_type = 'subscription' WHERE tier_type IS NULL;

-- ============================================
-- Verification Fees Table
-- Fee structure for function verification by level
-- ============================================
CREATE TABLE IF NOT EXISTS verification_fees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Verification level this fee applies to
    level VARCHAR(20) NOT NULL, -- 'basic', 'standard', 'full'

    -- Fee details
    price_cents INTEGER NOT NULL, -- Price in cents (e.g., 500 for $5)
    currency VARCHAR(3) DEFAULT 'USD',

    -- Is this tier currently active
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Optional: minimum plan required to access this tier
    min_plan VARCHAR(20), -- NULL = available to all plans

    -- Description for display
    description TEXT,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default verification fee structure
-- Level 1 (basic) is free to bootstrap marketplace
-- Level 2+ has fees ranging $5-25
INSERT INTO verification_fees (level, price_cents, description, min_plan) VALUES
    ('basic', 0, 'Basic verification - malware scan only (Free to bootstrap)', NULL),
    ('standard', 1500, 'Standard verification - malware + DRE + FXCERT ($15)', 'starter'),
    ('full', 2500, 'Full verification - all checks + manual review ($25)', 'pro')
ON CONFLICT DO NOTHING;

-- Indexes for verification_fees
CREATE INDEX idx_verification_fees_level ON verification_fees(level);
CREATE INDEX idx_verification_fees_active ON verification_fees(is_active) WHERE is_active = true;

-- ============================================
-- Function Verification Payments
-- Track payments for function verifications
-- ============================================
CREATE TABLE IF NOT EXISTS function_verification_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Reference to function and verification
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    verification_level VARCHAR(20) NOT NULL,

    -- Payment details
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'paid', 'refunded', 'failed'

    -- Stripe payment reference
    stripe_payment_intent_id VARCHAR(100),
    stripe_checkout_session_id VARCHAR(100),

    -- Payer
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    paid_by UUID REFERENCES users(id),

    -- Verification result reference
    verification_job_id UUID REFERENCES verification_jobs(id),

    -- Timing
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for function_verification_payments
CREATE INDEX idx_function_verification_payments_function_id ON function_verification_payments(function_id);
CREATE INDEX idx_function_verification_payments_tenant_id ON function_verification_payments(tenant_id);
CREATE INDEX idx_function_verification_payments_status ON function_verification_payments(status);
CREATE INDEX idx_function_verification_payments_stripe_pi ON function_verification_payments(stripe_payment_intent_id);

-- ============================================
-- Publisher Earnings Table
-- Track earnings from function sales
-- ============================================
CREATE TABLE IF NOT EXISTS publisher_earnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Publisher info
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    publisher_user_id UUID NOT NULL REFERENCES users(id),

    -- Function being sold
    function_id UUID REFERENCES registry_functions(id) ON DELETE SET NULL,
    function_name VARCHAR(255),

    -- Transaction details
    transaction_type VARCHAR(20) NOT NULL, -- 'sale', 'refund', 'payout'
    amount_cents INTEGER NOT NULL, -- Positive for earnings, negative for refunds
    currency VARCHAR(3) DEFAULT 'USD',

    -- Platform fee breakdown
    gross_amount_cents INTEGER NOT NULL,
    platform_fee_cents INTEGER NOT NULL DEFAULT 0,
    net_amount_cents INTEGER NOT NULL,
    platform_fee_percent NUMERIC(5,2) DEFAULT 15.00, -- 15% default platform cut

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'available', 'withdrawn', 'withheld'

    -- Stripe payout reference
    stripe_payout_id VARCHAR(100),

    -- Period for aggregation
    earned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    period_month INTEGER, -- 1-12
    period_year INTEGER, -- e.g., 2026

    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for publisher_earnings
CREATE INDEX idx_publisher_earnings_tenant_id ON publisher_earnings(tenant_id);
CREATE INDEX idx_publisher_earnings_user_id ON publisher_earnings(publisher_user_id);
CREATE INDEX idx_publisher_earnings_function_id ON publisher_earnings(function_id);
CREATE INDEX idx_publisher_earnings_status ON publisher_earnings(status);
CREATE INDEX idx_publisher_earnings_earned_at ON publisher_earnings(earned_at);
CREATE INDEX idx_publisher_earnings_period ON publisher_earnings(period_year, period_month);

-- ============================================
-- Agent Subscriptions Table
-- Track agent-based subscriptions ($10/agent/month)
-- ============================================
CREATE TABLE IF NOT EXISTS agent_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Agent info
    agent_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Plan details
    plan_name VARCHAR(50) NOT NULL DEFAULT 'per_agent', -- 'per_agent', 'unlimited'
    price_per_agent_cents INTEGER NOT NULL DEFAULT 1000, -- $10 in cents
    currency VARCHAR(3) DEFAULT 'USD',

    -- Agent limits
    max_agents INTEGER NOT NULL DEFAULT 1,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'suspended', 'cancelled'

    -- Billing period
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,

    -- Stripe subscription reference
    stripe_subscription_id VARCHAR(100),
    stripe_customer_id VARCHAR(100),

    -- Payment status
    last_payment_status VARCHAR(20), -- 'paid', 'failed', 'pending'
    last_payment_at TIMESTAMPTZ,

    -- Cancellation
    cancel_at_period_end BOOLEAN DEFAULT false,
    cancelled_at TIMESTAMPTZ,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for agent_subscriptions
CREATE INDEX idx_agent_subscriptions_tenant_id ON agent_subscriptions(tenant_id);
CREATE INDEX idx_agent_subscriptions_agent_id ON agent_subscriptions(agent_id);
CREATE INDEX idx_agent_subscriptions_status ON agent_subscriptions(status);
CREATE INDEX idx_agent_subscriptions_stripe_sub ON agent_subscriptions(stripe_subscription_id);

-- ============================================
-- Agent Usage Tracking
-- Track agent usage for billing purposes
-- ============================================
CREATE TABLE IF NOT EXISTS agent_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Agent and tenant
    agent_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    subscription_id UUID REFERENCES agent_subscriptions(id),

    -- Usage metrics
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,

    -- Counters
    total_calls INTEGER NOT NULL DEFAULT 0,
    total_executions INTEGER NOT NULL DEFAULT 0,
    total_errors INTEGER NOT NULL DEFAULT 0,
    total_latency_ms BIGINT NOT NULL DEFAULT 0,

    -- Cost calculation
    billable_calls INTEGER NOT NULL DEFAULT 0, -- Calls within quota
    overage_calls INTEGER NOT NULL DEFAULT 0, -- Calls exceeding quota
    estimated_cost_cents INTEGER NOT NULL DEFAULT 0,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'billed', 'disputed'

    -- Stripe invoice reference
    stripe_invoice_id VARCHAR(100),

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for agent_usage
CREATE INDEX idx_agent_usage_tenant_id ON agent_usage(tenant_id);
CREATE INDEX idx_agent_usage_agent_id ON agent_usage(agent_id);
CREATE INDEX idx_agent_usage_subscription_id ON agent_usage(subscription_id);
CREATE INDEX idx_agent_usage_period ON agent_usage(period_start, period_end);
CREATE INDEX idx_agent_usage_status ON agent_usage(status);

-- ============================================
-- Platform Fees Table
-- Track platform fees collected from marketplace
-- ============================================
-- 20260330000000_platform_fees created a different platform_fees schema (publish/commission audit).
-- CREATE TABLE IF NOT EXISTS would skip and leave missing columns (e.g. source_type). Preserve old data.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'platform_fees'
    ) AND EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'platform_fees' AND column_name = 'amount_usd'
    ) THEN
        ALTER TABLE platform_fees RENAME TO platform_fees_legacy_publish_audit;
        -- Index names are global; free idx_platform_fees_status for the new revenue platform_fees table.
        IF EXISTS (
            SELECT 1 FROM pg_class c
            JOIN pg_namespace n ON n.oid = c.relnamespace
            WHERE n.nspname = 'public' AND c.relkind = 'i' AND c.relname = 'idx_platform_fees_status'
        ) THEN
            ALTER INDEX idx_platform_fees_status RENAME TO idx_platform_fees_legacy_status;
        END IF;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS platform_fees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Fee type
    fee_type VARCHAR(20) NOT NULL, -- 'marketplace_sale', 'verification', 'agent_subscription', 'tier_upgrade'

    -- Reference to source transaction
    source_transaction_id UUID, -- Links to publisher_earnings, function_verification_payments, etc.
    source_type VARCHAR(50), -- 'publisher_earnings', 'function_verification_payment', 'agent_subscription'

    -- Amount details
    gross_amount_cents INTEGER NOT NULL,
    platform_fee_cents INTEGER NOT NULL,
    net_amount_cents INTEGER NOT NULL,
    platform_fee_percent NUMERIC(5,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',

    -- Parties involved
    tenant_id UUID REFERENCES tenants(id),
    user_id UUID REFERENCES users(id),
    function_id UUID REFERENCES registry_functions(id),
    agent_id UUID,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'collected', -- 'collected', 'refunded', 'disputed', 'paid_out'

    -- Stripe payout
    stripe_transfer_id VARCHAR(100),
    paid_out_at TIMESTAMPTZ,

    -- Period for reporting
    period_month INTEGER,
    period_year INTEGER,

    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for platform_fees
CREATE INDEX idx_platform_fees_type ON platform_fees(fee_type);
CREATE INDEX idx_platform_fees_source ON platform_fees(source_type, source_transaction_id);
CREATE INDEX idx_platform_fees_tenant_id ON platform_fees(tenant_id);
CREATE INDEX idx_platform_fees_period ON platform_fees(period_year, period_month);
CREATE INDEX idx_platform_fees_status ON platform_fees(status);
CREATE INDEX idx_platform_fees_created_at ON platform_fees(created_at);

-- ============================================
-- Add columns to existing tables for revenue system
-- ============================================

-- Add publisher info to registry_functions for earnings tracking
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS is_marketplace_listing BOOLEAN DEFAULT false;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS listing_price_cents INTEGER;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS publisher_tenant_id UUID REFERENCES tenants(id);
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS publisher_user_id UUID REFERENCES users(id);

-- Add marketplace columns to registry_function_versions if needed
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS version_sales_count INTEGER DEFAULT 0;
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS version_total_earnings_cents INTEGER DEFAULT 0;

-- ============================================
-- Update pricing_tiers with new Moat tiers
-- ============================================
UPDATE pricing_tiers SET
    tier_type = 'subscription',
    max_agents = 1,
    max_functions = 10,
    max_executions_per_month = 1000,
    trial_days = 14
WHERE name ILIKE '%starter%' OR name ILIKE '%free%';

COMMIT;
