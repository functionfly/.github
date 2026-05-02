-- Migration: Remove Trust API billing infrastructure
-- Reverses: 20260412190001_trust_api_billing

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_tier_pricing_timestamp ON trust_api_tier_pricing;
DROP TRIGGER IF EXISTS trigger_update_partner_billing_usage_timestamp ON trust_api_partner_billing_usage;

-- Drop helper functions
DROP FUNCTION IF EXISTS update_trust_api_tier_pricing_updated_at();
DROP FUNCTION IF EXISTS update_trust_partner_billing_usage_updated_at();

-- Drop RLS policies (must drop policies before disabling RLS)
ALTER TABLE trust_api_tier_pricing DISABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_partner_billing_usage DISABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_billing_records DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tier_pricing_read ON trust_api_tier_pricing;
DROP POLICY IF EXISTS partner_billing_usage_read_own ON trust_api_partner_billing_usage;
DROP POLICY IF EXISTS billing_records_read_own ON trust_api_billing_records;

-- Drop billing tables
DROP TABLE IF EXISTS trust_api_billing_records;
DROP TABLE IF EXISTS trust_api_partner_billing_usage;
DROP TABLE IF EXISTS trust_api_tier_pricing;

-- Remove billing columns from trust_api_partners
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS billing_status;
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS stripe_customer_id;
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS stripe_subscription_id;
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS stripe_price_id;
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS is_founder_mode;
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS founder_mode_started_at;
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS founder_mode_ends_at;
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS usage_threshold;
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS current_overage_usage;
ALTER TABLE trust_api_partners DROP COLUMN IF EXISTS total_billed_overage;

-- Drop indexes
DROP INDEX IF EXISTS idx_trust_partners_billing_status;
DROP INDEX IF EXISTS idx_trust_partners_stripe_price;