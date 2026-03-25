-- Migration: 20260402000000_revenue_phase1
-- Description: Revenue System Phase 1 - Trust Layer Monetization (Rollback)
-- Created: 2026-04-02

BEGIN;

-- Remove publisher info from registry_functions
ALTER TABLE registry_functions DROP COLUMN IF EXISTS is_marketplace_listing;
ALTER TABLE registry_functions DROP COLUMN IF EXISTS listing_price_cents;
ALTER TABLE registry_functions DROP COLUMN IF EXISTS publisher_tenant_id;
ALTER TABLE registry_functions DROP COLUMN IF EXISTS publisher_user_id;

-- Remove marketplace columns from registry_function_versions
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS version_sales_count;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS version_total_earnings_cents;

-- Remove pricing tier extensions
ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS tier_type;
ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS stripe_price_id;
ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS trial_days;
ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS max_agents;
ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS max_functions;
ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS max_executions_per_month;

-- Drop tables in reverse order of creation (respecting foreign key dependencies)
DROP TABLE IF EXISTS platform_fees;
-- Restore legacy platform_fees from 20260330000000 if this migration renamed it aside
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'platform_fees_legacy_publish_audit'
    ) THEN
        ALTER TABLE platform_fees_legacy_publish_audit RENAME TO platform_fees;
        IF EXISTS (
            SELECT 1 FROM pg_class c
            JOIN pg_namespace n ON n.oid = c.relnamespace
            WHERE n.nspname = 'public' AND c.relkind = 'i' AND c.relname = 'idx_platform_fees_legacy_status'
        ) THEN
            ALTER INDEX idx_platform_fees_legacy_status RENAME TO idx_platform_fees_status;
        END IF;
    END IF;
END $$;
DROP TABLE IF EXISTS agent_usage;
DROP TABLE IF EXISTS agent_subscriptions;
DROP TABLE IF EXISTS publisher_earnings;
DROP TABLE IF EXISTS function_verification_payments;
DROP TABLE IF EXISTS verification_fees;

COMMIT;
