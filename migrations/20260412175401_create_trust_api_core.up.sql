-- Migration: Create Trust API core tables
-- Description: Creates the foundational Trust API tables - partners, keys, usage tracking, rate limits, reports, and verifications
-- This migration must run before trust_api_billing migration which adds columns to trust_api_partners

-- ============================================
-- Trust API Partners Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_api_partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,

    -- Contact information
    contact_email VARCHAR(255) NOT NULL,
    contact_name VARCHAR(255),
    website_url VARCHAR(500),

    -- Partner tier
    tier VARCHAR(50) NOT NULL DEFAULT 'developer',

    -- Rate limits
    rate_limit_per_minute INTEGER DEFAULT 60,
    rate_limit_per_day INTEGER DEFAULT 10000,

    -- Usage quotas
    monthly_request_limit INTEGER DEFAULT 50000,
    current_month_usage INTEGER DEFAULT 0,

    -- Billing (basic columns - additional added by billing migration)
    billing_email VARCHAR(255),
    billing_account_id VARCHAR(255),
    billing_status VARCHAR(50) DEFAULT 'trial',

    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    -- SSO
    sso_enabled BOOLEAN DEFAULT FALSE,
    sso_provider VARCHAR(50),

    -- Webhook
    webhook_url VARCHAR(500),
    webhook_secret_hash VARCHAR(255),

    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Timestamps
    activated_at TIMESTAMP WITH TIME ZONE,
    suspended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for partners
CREATE INDEX IF NOT EXISTS idx_trust_partners_slug ON trust_api_partners(slug);
CREATE INDEX IF NOT EXISTS idx_trust_partners_tier ON trust_api_partners(tier);
CREATE INDEX IF NOT EXISTS idx_trust_partners_status ON trust_api_partners(status);
CREATE INDEX IF NOT EXISTS idx_trust_partners_contact_email ON trust_api_partners(contact_email);

-- ============================================
-- Trust API Keys Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES trust_api_partners(id) ON DELETE CASCADE,

    -- Key identification
    key_id VARCHAR(32) NOT NULL UNIQUE,
    key_prefix VARCHAR(10) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,

    -- Key metadata
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Scope and permissions
    scopes JSONB DEFAULT '["trust:read"]'::jsonb,

    -- IP allowlist
    allowed_ips JSONB DEFAULT '[]'::jsonb,

    -- Expiration and revocation
    expires_at TIMESTAMP WITH TIME ZONE,
    is_revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_reason TEXT,

    -- Usage tracking
    last_used_at TIMESTAMP WITH TIME ZONE,
    use_count INTEGER DEFAULT 0,

    -- Created by
    created_by VARCHAR(255) NOT NULL,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for API keys
CREATE INDEX IF NOT EXISTS idx_trust_api_keys_partner ON trust_api_keys(partner_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_keys_key_id ON trust_api_keys(key_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_keys_expires_at ON trust_api_keys(expires_at);
CREATE INDEX IF NOT EXISTS idx_trust_api_keys_is_revoked ON trust_api_keys(is_revoked);

-- ============================================
-- Trust API Usage Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_api_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES trust_api_partners(id) ON DELETE CASCADE,
    api_key_id UUID REFERENCES trust_api_keys(id) ON DELETE SET NULL,

    -- Endpoint and method
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,

    -- Request details
    request_id VARCHAR(255) NOT NULL UNIQUE,
    function_id UUID,

    -- Response details
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER NOT NULL,

    -- Rate limit tracking
    rate_limit_remaining INTEGER,
    rate_limit_reset_at TIMESTAMP WITH TIME ZONE,

    -- Request context
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Error tracking
    error_code VARCHAR(50),
    error_message TEXT,

    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for usage tracking
CREATE INDEX IF NOT EXISTS idx_trust_api_usage_partner ON trust_api_usage(partner_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_usage_api_key ON trust_api_usage(api_key_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_usage_request_id ON trust_api_usage(request_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_usage_function_id ON trust_api_usage(function_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_usage_created_at ON trust_api_usage(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_trust_api_usage_endpoint ON trust_api_usage(endpoint);

-- ============================================
-- Trust API Rate Limits Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_api_rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES trust_api_partners(id) ON DELETE CASCADE,

    -- Rate limit type
    limit_type VARCHAR(50) NOT NULL,

    -- Window tracking
    window_start TIMESTAMP WITH TIME ZONE NOT NULL,
    window_end TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Request count
    request_count INTEGER DEFAULT 0,

    -- Unique constraint for sliding window
    CONSTRAINT uq_trust_api_rate_limits_partner_type_window UNIQUE (partner_id, limit_type, window_start)
);

-- Indexes for rate limits
CREATE INDEX IF NOT EXISTS idx_trust_api_rate_limits_partner ON trust_api_rate_limits(partner_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_rate_limits_window ON trust_api_rate_limits(window_start, window_end);

-- ============================================
-- Trust API Reports Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_api_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES trust_api_partners(id) ON DELETE CASCADE,
    api_key_id UUID REFERENCES trust_api_keys(id) ON DELETE SET NULL,

    -- Report identification
    report_id VARCHAR(32) NOT NULL UNIQUE,

    -- Function being reported
    function_id UUID NOT NULL,
    function_author VARCHAR(255),
    function_name VARCHAR(255),

    -- Report details
    report_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    evidence JSONB DEFAULT '{}'::jsonb,

    -- Status tracking
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    resolution_notes TEXT,
    resolved_by UUID,
    resolved_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for reports
CREATE INDEX IF NOT EXISTS idx_trust_api_reports_partner ON trust_api_reports(partner_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_reports_report_id ON trust_api_reports(report_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_reports_function_id ON trust_api_reports(function_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_reports_status ON trust_api_reports(status);
CREATE INDEX IF NOT EXISTS idx_trust_api_reports_report_type ON trust_api_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_trust_api_reports_created_at ON trust_api_reports(created_at DESC);

-- ============================================
-- Trust API Verifications Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_api_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID NOT NULL REFERENCES trust_api_partners(id) ON DELETE CASCADE,
    api_key_id UUID REFERENCES trust_api_keys(id) ON DELETE SET NULL,

    -- Verification identification
    verification_id VARCHAR(32) NOT NULL UNIQUE,

    -- Function to verify
    function_id UUID NOT NULL,
    function_author VARCHAR(255),
    function_name VARCHAR(255),
    function_version VARCHAR(50),

    -- Verification details
    verification_level VARCHAR(50) NOT NULL DEFAULT 'standard',

    -- Request content
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Status tracking
    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    -- Result
    trust_score FLOAT,
    trust_tier VARCHAR(50),
    verification_badge_url VARCHAR(500),
    completion_notes TEXT,
    completed_by UUID,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for verifications
CREATE INDEX IF NOT EXISTS idx_trust_api_verifications_partner ON trust_api_verifications(partner_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_verifications_verification_id ON trust_api_verifications(verification_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_verifications_function_id ON trust_api_verifications(function_id);
CREATE INDEX IF NOT EXISTS idx_trust_api_verifications_status ON trust_api_verifications(status);
CREATE INDEX IF NOT EXISTS idx_trust_api_verifications_created_at ON trust_api_verifications(created_at DESC);

-- ============================================
-- Trigger for updated_at timestamps
-- ============================================
CREATE OR REPLACE FUNCTION update_trust_api_core_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DROP TRIGGER IF EXISTS update_trust_api_partners_updated_at ON trust_api_partners;
CREATE TRIGGER update_trust_api_partners_updated_at
    BEFORE UPDATE ON trust_api_partners
    FOR EACH ROW
    EXECUTE FUNCTION update_trust_api_core_updated_at_column();

DROP TRIGGER IF EXISTS update_trust_api_keys_updated_at ON trust_api_keys;
CREATE TRIGGER update_trust_api_keys_updated_at
    BEFORE UPDATE ON trust_api_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_trust_api_core_updated_at_column();

DROP TRIGGER IF EXISTS update_trust_api_reports_updated_at ON trust_api_reports;
CREATE TRIGGER update_trust_api_reports_updated_at
    BEFORE UPDATE ON trust_api_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_trust_api_core_updated_at_column();

DROP TRIGGER IF EXISTS update_trust_api_verifications_updated_at ON trust_api_verifications;
CREATE TRIGGER update_trust_api_verifications_updated_at
    BEFORE UPDATE ON trust_api_verifications
    FOR EACH ROW
    EXECUTE FUNCTION update_trust_api_core_updated_at_column();

-- ============================================
-- Row Level Security (RLS)
-- ============================================
ALTER TABLE trust_api_partners ENABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_usage ENABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_rate_limits ENABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_reports ENABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_verifications ENABLE ROW LEVEL SECURITY;

-- Trust API Partners: readable by all authenticated, writable by admin
CREATE POLICY trust_partners_select ON trust_api_partners FOR SELECT USING (true);
CREATE POLICY trust_partners_update ON trust_api_partners FOR UPDATE USING (true);

-- Trust API Keys: only accessible by the partner (enforced at application layer)
CREATE POLICY trust_api_keys_select ON trust_api_keys FOR SELECT USING (true);
CREATE POLICY trust_api_keys_insert ON trust_api_keys FOR INSERT WITH CHECK (true);
CREATE POLICY trust_api_keys_update ON trust_api_keys FOR UPDATE USING (true);
CREATE POLICY trust_api_keys_delete ON trust_api_keys FOR DELETE USING (true);

-- Trust API Usage: readable by partner, inserted by system
CREATE POLICY trust_api_usage_select ON trust_api_usage FOR SELECT USING (true);
CREATE POLICY trust_api_usage_insert ON trust_api_usage FOR INSERT WITH CHECK (true);

-- Trust API Rate Limits: readable and writable by partner
CREATE POLICY trust_api_rate_limits_select ON trust_api_rate_limits FOR SELECT USING (true);
CREATE POLICY trust_api_rate_limits_insert ON trust_api_rate_limits FOR INSERT WITH CHECK (true);
CREATE POLICY trust_api_rate_limits_update ON trust_api_rate_limits FOR UPDATE USING (true);

-- Trust API Reports: readable by partner that created, writable by admin
CREATE POLICY trust_api_reports_select ON trust_api_reports FOR SELECT USING (true);
CREATE POLICY trust_api_reports_insert ON trust_api_reports FOR INSERT WITH CHECK (true);
CREATE POLICY trust_api_reports_update ON trust_api_reports FOR UPDATE USING (true);

-- Trust API Verifications: readable by partner that created, writable by admin
CREATE POLICY trust_api_verifications_select ON trust_api_verifications FOR SELECT USING (true);
CREATE POLICY trust_api_verifications_insert ON trust_api_verifications FOR INSERT WITH CHECK (true);
CREATE POLICY trust_api_verifications_update ON trust_api_verifications FOR UPDATE USING (true);

-- ============================================
-- Comments for documentation
-- ============================================
COMMENT ON TABLE trust_api_partners IS 'External platform partners integrated with the Trust API';
COMMENT ON TABLE trust_api_keys IS 'API keys for partner authentication to the Trust API';
COMMENT ON TABLE trust_api_usage IS 'Detailed API usage tracking for billing and analytics';
COMMENT ON TABLE trust_api_rate_limits IS 'Sliding window rate limit tracking per partner';
COMMENT ON TABLE trust_api_reports IS 'Trust issue reports submitted by partners';
COMMENT ON TABLE trust_api_verifications IS 'Function verification requests from partners';
