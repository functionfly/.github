-- Migration: Add subscription performance indexes for billing queries
-- Created: 2026-06-18

-- ============================================================
-- Indexes for subscription lookups
-- ============================================================

-- Composite index on subscriptions for tenant and status lookups
-- Optimizes queries that filter by tenant_id and status together
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_status
    ON subscriptions(tenant_id, status);

-- Index for bundle_subscriptions tenant and status lookups
CREATE INDEX IF NOT EXISTS idx_bundle_subscriptions_tenant_status
    ON bundle_subscriptions(tenant_id, status);

-- ============================================================
-- Indexes for affiliate referral queries (N+1 fix)
-- ============================================================

-- Composite index for efficient affiliate referral lookups by code
CREATE INDEX IF NOT EXISTS idx_affiliate_referrals_code_created
    ON affiliate_referrals(affiliate_code_id, created_at DESC);

-- Index for publisher referral summary queries
CREATE INDEX IF NOT EXISTS idx_affiliate_codes_publisher_active
    ON affiliate_codes(publisher_id, is_active, created_at DESC);

-- ============================================================
-- Index comments for documentation
-- ============================================================
COMMENT ON INDEX idx_subscriptions_tenant_status IS
    'Optimizes subscription lookups by tenant and status for billing queries';
COMMENT ON INDEX idx_bundle_subscriptions_tenant_status IS
    'Optimizes bundle subscription lookups by tenant and status';
COMMENT ON INDEX idx_affiliate_referrals_code_created IS
    'Optimizes affiliate referral queries by code ID for N+1 fix';
COMMENT ON INDEX idx_affiliate_codes_publisher_active IS
    'Optimizes affiliate code lookups by publisher for referral listing';
