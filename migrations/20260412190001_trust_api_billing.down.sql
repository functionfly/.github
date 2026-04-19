-- Down migration: Trust API Billing Infrastructure

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_tier_pricing_timestamp ON trust_api_tier_pricing;
DROP TRIGGER IF EXISTS trigger_update_partner_billing_usage_timestamp ON trust_api_partner_billing_usage;
DROP FUNCTION IF EXISTS update_trust_api_tier_pricing_updated_at();
DROP FUNCTION IF EXISTS update_trust_partner_billing_usage_updated_at();

-- Drop RLS policies
DROP POLICY IF EXISTS tier_pricing_read ON trust_api_tier_pricing;
DROP POLICY IF EXISTS partner_billing_usage_read_own ON trust_api_partner_billing_usage;
DROP POLICY IF EXISTS billing_records_read_own ON trust_api_billing_records;

-- Disable RLS
ALTER TABLE trust_api_tier_pricing DISABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_partner_billing_usage DISABLE ROW LEVEL SECURITY;
ALTER TABLE trust_api_billing_records DISABLE ROW LEVEL SECURITY;

-- Drop tables
DROP TABLE IF EXISTS trust_api_billing_records;
DROP TABLE IF EXISTS trust_api_partner_billing_usage;
DROP TABLE IF EXISTS trust_api_tier_pricing;

-- Remove columns from trust_api_partners
ALTER TABLE trust_api_partners
    DROP COLUMN IF EXISTS billing_status,
    DROP COLUMN IF EXISTS stripe_customer_id,
    DROP COLUMN IF EXISTS stripe_subscription_id,
    DROP COLUMN IF EXISTS stripe_price_id,
    DROP COLUMN IF EXISTS is_founder_mode,
    DROP COLUMN IF EXISTS founder_mode_started_at,
    DROP COLUMN IF EXISTS founder_mode_ends_at,
    DROP COLUMN IF EXISTS usage_threshold,
    DROP COLUMN IF EXISTS current_overage_usage,
    DROP COLUMN IF EXISTS total_billed_overage;
