-- Remove annual pricing columns from pricing_tiers
ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS billing_cycle;
ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS annual_price_cents;
ALTER TABLE pricing_tiers DROP COLUMN IF EXISTS stripe_price_id_annual;

-- Remove billing_cycle from subscriptions
ALTER TABLE subscriptions DROP COLUMN IF EXISTS billing_cycle;