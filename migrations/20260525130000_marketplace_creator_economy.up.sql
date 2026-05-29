-- Marketplace creator economy: subscription plans, customer subscriptions, payout requests

BEGIN;

CREATE TABLE IF NOT EXISTS marketplace_subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(12, 4) NOT NULL DEFAULT 0,
    features JSONB NOT NULL DEFAULT '[]'::jsonb,
    billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT marketplace_subscription_plans_billing_cycle_check
        CHECK (billing_cycle IN ('monthly', 'quarterly', 'annual'))
);

CREATE INDEX IF NOT EXISTS idx_marketplace_subscription_plans_tenant
    ON marketplace_subscription_plans (tenant_id);

CREATE TABLE IF NOT EXISTS marketplace_plan_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID REFERENCES marketplace_subscription_plans(id) ON DELETE SET NULL,
    creator_tenant_id UUID NOT NULL,
    subscriber_tenant_id UUID,
    subscriber_user_id UUID,
    subscriber_name VARCHAR(255) NOT NULL DEFAULT '',
    subscriber_email VARCHAR(255) NOT NULL DEFAULT '',
    plan_name VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    amount DECIMAL(12, 4) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_cycle VARCHAR(20) NOT NULL DEFAULT 'monthly',
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '30 days'),
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT marketplace_plan_subscriptions_status_check
        CHECK (status IN ('active', 'cancelled', 'past_due', 'trialing'))
);

CREATE INDEX IF NOT EXISTS idx_marketplace_plan_subscriptions_creator
    ON marketplace_plan_subscriptions (creator_tenant_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_plan_subscriptions_status
    ON marketplace_plan_subscriptions (creator_tenant_id, status);

CREATE TABLE IF NOT EXISTS marketplace_payout_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    amount DECIMAL(12, 4) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    requested_by_user_id UUID,
    notes TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT marketplace_payout_requests_status_check
        CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_marketplace_payout_requests_tenant
    ON marketplace_payout_requests (tenant_id, created_at DESC);

COMMENT ON TABLE marketplace_subscription_plans IS 'Creator-defined subscription plans for marketplace functions';
COMMENT ON TABLE marketplace_plan_subscriptions IS 'Customer subscriptions to creator marketplace plans';
COMMENT ON TABLE marketplace_payout_requests IS 'Creator royalty payout requests';

COMMIT;
