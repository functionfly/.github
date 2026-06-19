-- Migration: Remove subscription performance indexes
-- Created: 2026-06-18

-- ============================================================
-- Drop subscription indexes
-- ============================================================

DROP INDEX IF EXISTS idx_subscriptions_tenant_status;
DROP INDEX IF EXISTS idx_bundle_subscriptions_tenant_status;
DROP INDEX IF EXISTS idx_affiliate_referrals_code_created;
DROP INDEX IF EXISTS idx_affiliate_codes_publisher_active;
