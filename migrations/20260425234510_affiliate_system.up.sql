-- Affiliate / Referral Commission System Tables
-- Supports promo codes with referral commission tracking

-- Affiliate Codes (promo codes for affiliates)
CREATE TABLE IF NOT EXISTS affiliate_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    publisher_id UUID NOT NULL,
    tenant_id UUID, -- Optional: tie to a specific tenant/publisher account

    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Commission structure
    commission_type VARCHAR(20) NOT NULL DEFAULT 'percent', -- 'percent' or 'fixed'
    commission_value DECIMAL(10, 4) NOT NULL, -- e.g., 20.00 for 20% or $20 fixed

    -- Limits
    max_commissions INT, -- Max total commissions payout (NULL = unlimited)
    max_referrals INT, -- Max number of referrals (NULL = unlimited)

    -- Current counts
    total_referrals INT NOT NULL DEFAULT 0,
    total_commissions INT NOT NULL DEFAULT 0,
    pending_commissions INT NOT NULL DEFAULT 0,

    -- Earnings tracking (in cents for precision)
    pending_earnings_cents BIGINT NOT NULL DEFAULT 0,
    total_earnings_cents BIGINT NOT NULL DEFAULT 0,
    paid_out_earnings_cents BIGINT NOT NULL DEFAULT 0,

    -- Validity
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Metadata
    utm_source VARCHAR(255),
    utm_campaign VARCHAR(255),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_affiliate_codes_publisher ON affiliate_codes(publisher_id);
CREATE INDEX IF NOT EXISTS idx_affiliate_codes_code ON affiliate_codes(UPPER(code));
CREATE INDEX IF NOT EXISTS idx_affiliate_codes_active ON affiliate_codes(is_active);

-- Affiliate Referrals (tracks referred tenants)
CREATE TABLE IF NOT EXISTS affiliate_referrals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    affiliate_code_id UUID NOT NULL REFERENCES affiliate_codes(id),
    referred_tenant_id UUID NOT NULL,
    subscription_id UUID, -- Filled when they subscribe

    -- Attribution data
    utm_source VARCHAR(255),
    utm_campaign VARCHAR(255),
    utm_content VARCHAR(255),
    utm_term VARCHAR(255),

    ip_address INET,
    user_agent TEXT,

    -- Tracking status
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'converted', 'qualified', 'canceled'

    -- When the referral was made
    referred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- When they subscribed (if they did)
    converted_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_affiliate_referrals_code ON affiliate_referrals(affiliate_code_id);
CREATE INDEX IF NOT EXISTS idx_affiliate_referrals_tenant ON affiliate_referrals(referred_tenant_id);
CREATE INDEX IF NOT EXISTS idx_affiliate_referrals_status ON affiliate_referrals(status);
CREATE INDEX IF NOT EXISTS idx_affiliate_referrals_referred_at ON affiliate_referrals(referred_at);

-- Affiliate Commissions (earned commissions)
CREATE TABLE IF NOT EXISTS affiliate_commissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    affiliate_code_id UUID NOT NULL REFERENCES affiliate_codes(id),
    referral_id UUID NOT NULL REFERENCES affiliate_referrals(id),

    -- Commission details
    commission_type VARCHAR(20) NOT NULL, -- 'percent' or 'fixed'
    commission_value DECIMAL(10, 4) NOT NULL,

    -- What the commission is based on
    base_amount_cents BIGINT NOT NULL, -- Subscription amount in cents
    base_amount_usd DECIMAL(14, 4) NOT NULL,

    -- Calculated commission
    commission_cents BIGINT NOT NULL,
    commission_usd DECIMAL(14, 4) NOT NULL,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'paid', 'canceled'

    -- For tracking payment
    paid_at TIMESTAMPTZ,
    payment_batch_id UUID,
    payment_batch VARCHAR(100), -- Human-readable batch ID

    -- Associated subscription for commission period
    subscription_id UUID,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_affiliate_commissions_code ON affiliate_commissions(affiliate_code_id);
CREATE INDEX IF NOT EXISTS idx_affiliate_commissions_referral ON affiliate_commissions(referral_id);
CREATE INDEX IF NOT EXISTS idx_affiliate_commissions_status ON affiliate_commissions(status);
CREATE INDEX IF NOT EXISTS idx_affiliate_commissions_paid ON affiliate_commissions(paid_at);
CREATE INDEX IF NOT EXISTS idx_affiliate_commissions_created ON affiliate_commissions(created_at);

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE TRIGGER update_affiliate_codes_updated_at
    BEFORE UPDATE ON affiliate_codes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_affiliate_referrals_updated_at
    BEFORE UPDATE ON affiliate_referrals
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_affiliate_commissions_updated_at
    BEFORE UPDATE ON affiliate_commissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();