-- Migration: Trust API Billing Infrastructure
-- Adds billing support for partner monetization

-- ============================================
-- 1. Add billing columns to trust_api_partners
-- ============================================
ALTER TABLE trust_api_partners
    ADD COLUMN IF NOT EXISTS billing_status VARCHAR(50) DEFAULT 'trial',
    ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS stripe_price_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS is_founder_mode BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS founder_mode_started_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS founder_mode_ends_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS usage_threshold INTEGER DEFAULT 100000,
    ADD COLUMN IF NOT EXISTS current_overage_usage INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_billed_overage INTEGER DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_trust_partners_billing_status ON trust_api_partners(billing_status);
CREATE INDEX IF NOT EXISTS idx_trust_partners_stripe_price ON trust_api_partners(stripe_price_id);

-- ============================================
-- 2. Create tier pricing table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_api_tier_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier VARCHAR(50) NOT NULL UNIQUE,
    monthly_price_cents INTEGER NOT NULL DEFAULT 0,
    included_requests INTEGER NOT NULL DEFAULT 0,
    overage_price_per_1000 INTEGER NOT NULL DEFAULT 0,
    has_overage_billing BOOLEAN DEFAULT FALSE,
    stripe_price_id VARCHAR(255),
    rate_limit_per_minute INTEGER DEFAULT 60,
    rate_limit_per_day INTEGER DEFAULT 10000,
    monthly_request_limit INTEGER DEFAULT 50000,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert default pricing tiers
INSERT INTO trust_api_tier_pricing (
    tier, monthly_price_cents, included_requests, overage_price_per_1000,
    has_overage_billing, stripe_price_id, rate_limit_per_minute,
    rate_limit_per_day, monthly_request_limit, description, is_active
) VALUES
    ('developer', 0, 50000, 0, FALSE, NULL, 60, 10000, 50000,
     'Free tier for developers - 50K requests/month, hard limit', TRUE)
ON CONFLICT (tier) DO NOTHING;

INSERT INTO trust_api_tier_pricing (
    tier, monthly_price_cents, included_requests, overage_price_per_1000,
    has_overage_billing, stripe_price_id, rate_limit_per_minute,
    rate_limit_per_day, monthly_request_limit, description, is_active
) VALUES
    ('startup', 4900, 500000, 5, TRUE, NULL, 300, 100000, 500000,
     'Startup tier - $49/mo for 500K requests, $0.005 per overage', TRUE)
ON CONFLICT (tier) DO NOTHING;

INSERT INTO trust_api_tier_pricing (
    tier, monthly_price_cents, included_requests, overage_price_per_1000,
    has_overage_billing, stripe_price_id, rate_limit_per_minute,
    rate_limit_per_day, monthly_request_limit, description, is_active
) VALUES
    ('business', 19900, 2000000, 3, TRUE, NULL, 1000, 500000, 2000000,
     'Business tier - $199/mo for 2M requests, $0.003 per overage', TRUE)
ON CONFLICT (tier) DO NOTHING;

INSERT INTO trust_api_tier_pricing (
    tier, monthly_price_cents, included_requests, overage_price_per_1000,
    has_overage_billing, stripe_price_id, rate_limit_per_minute,
    rate_limit_per_day, monthly_request_limit, description, is_active
) VALUES
    ('enterprise', 0, 0, 0, FALSE, NULL, 10000, 10000000, 100000000,
     'Enterprise tier - Custom pricing, contact sales', TRUE)
ON CONFLICT (tier) DO NOTHING;

-- ============================================
-- 3. Create billing usage tracking table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_api_partner_billing_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL UNIQUE REFERENCES trust_api_partners(id) ON DELETE CASCADE,
    billing_period_start TIMESTAMP WITH TIME ZONE,
    billing_period_end TIMESTAMP WITH TIME ZONE,
    requests_this_period INTEGER DEFAULT 0,
    overages_this_period INTEGER DEFAULT 0,
    last_reported_at TIMESTAMP WITH TIME ZONE,
    last_reported_usage INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_trust_billing_usage_partner ON trust_api_partner_billing_usage(partner_id);
CREATE INDEX IF NOT EXISTS idx_trust_billing_usage_period ON trust_api_partner_billing_usage(billing_period_start, billing_period_end);

-- ============================================
-- 4. Create billing records table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_api_billing_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES trust_api_partners(id) ON DELETE CASCADE,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    base_requests INTEGER DEFAULT 0,
    overage_requests INTEGER DEFAULT 0,
    total_requests INTEGER DEFAULT 0,
    base_charge_cents INTEGER DEFAULT 0,
    overage_charge_cents INTEGER DEFAULT 0,
    total_charge_cents INTEGER DEFAULT 0,
    stripe_invoice_id VARCHAR(255),
    stripe_payment_status VARCHAR(50) DEFAULT 'pending',
    status VARCHAR(50) DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_trust_billing_records_partner ON trust_api_billing_records(partner_id);
CREATE INDEX IF NOT EXISTS idx_trust_billing_records_period ON trust_api_billing_records(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_trust_billing_records_status ON trust_api_billing_records(status);
CREATE INDEX IF NOT EXISTS idx_trust_billing_records_stripe_invoice ON trust_api_billing_records(stripe_invoice_id);

-- ============================================
-- 5. Create trigger to update timestamps
-- ============================================
CREATE OR REPLACE FUNCTION update_trust_api_tier_pricing_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_tier_pricing_timestamp ON trust_api_tier_pricing;
CREATE TRIGGER trigger_update_tier_pricing_timestamp
    BEFORE UPDATE ON trust_api_tier_pricing
    FOR EACH ROW
    EXECUTE FUNCTION update_trust_api_tier_pricing_updated_at();

CREATE OR REPLACE FUNCTION update_trust_partner_billing_usage_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_partner_billing_usage_timestamp ON trust_api_partner_billing_usage;
CREATE TRIGGER trigger_update_partner_billing_usage_timestamp
    BEFORE UPDATE ON trust_api_partner_billing_usage
    FOR EACH ROW
    EXECUTE FUNCTION update_trust_partner_billing_usage_updated_at();

-- ============================================
-- 6. Create RLS policies for billing tables
-- ============================================
ALTER TABLE trust_api_tier_pricing ENABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_partner_billing_usage ENABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_billing_records ENABLE ROW LEVEL SECURITY;

-- Tier pricing: readable by all active partners
CREATE POLICY tier_pricing_read ON trust_api_tier_pricing
    FOR SELECT USING (is_active = TRUE);

-- Partner billing usage: only accessible by the partner (via application layer)
CREATE POLICY partner_billing_usage_read_own ON trust_api_partner_billing_usage
    FOR SELECT USING (true); -- Application layer enforces access

-- Billing records: accessible by partner
CREATE POLICY billing_records_read_own ON trust_api_billing_records
    FOR SELECT USING (true); -- Application layer enforces access

-- ============================================
-- 7. Add comments
-- ============================================
COMMENT ON TABLE trust_api_tier_pricing IS 'Pricing configuration for Trust API partner tiers';
COMMENT ON TABLE trust_api_partner_billing_usage IS 'Real-time usage tracking for partner billing';
COMMENT ON TABLE trust_api_billing_records IS 'Historical billing records for partner invoicing';
