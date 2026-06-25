-- Add stripe_metered_price_id column to pricing_tiers for metered billing support
-- This column stores the Stripe price ID for metered (usage-based) billing

ALTER TABLE pricing_tiers
ADD COLUMN IF NOT EXISTS stripe_metered_price_id VARCHAR(255);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_pricing_tiers_stripe_metered_price_id
ON pricing_tiers(stripe_metered_price_id)
WHERE stripe_metered_price_id IS NOT NULL;
