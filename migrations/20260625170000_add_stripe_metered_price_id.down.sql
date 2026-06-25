-- Remove stripe_metered_price_id column from pricing_tiers

ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS stripe_metered_price_id;

DROP INDEX IF EXISTS idx_pricing_tiers_stripe_metered_price_id;
