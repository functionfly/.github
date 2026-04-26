-- Add annual pricing columns to pricing_tiers
ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS billing_cycle VARCHAR(10) DEFAULT 'monthly';
ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS annual_price_cents INTEGER;
ALTER TABLE pricing_tiers ADD COLUMN IF NOT EXISTS stripe_price_id_annual VARCHAR(255);

-- Add billing_cycle to subscriptions
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS billing_cycle VARCHAR(10) DEFAULT 'monthly';

-- Update existing pricing tiers with annual prices (20% discount for monthly plans)
-- Free tier: $0 annually
UPDATE pricing_tiers SET annual_price_cents = 0 WHERE name = 'Free' AND price_cents = 0;

-- Starter: $29/mo * 12 * 0.8 = $278.40 (~$27840 cents)
UPDATE pricing_tiers SET annual_price_cents = 27840 WHERE name = 'Starter' AND price_cents = 2900;

-- Professional: $99/mo * 12 * 0.8 = $950.40 (~$95040 cents)
UPDATE pricing_tiers SET annual_price_cents = 95040 WHERE name = 'Professional' AND price_cents = 9900;

-- Enterprise: Custom pricing (leave annual_price_cents as NULL)
-- Set billing_cycle for all existing tiers
UPDATE pricing_tiers SET billing_cycle = 'monthly' WHERE billing_cycle IS NULL;
UPDATE subscriptions SET billing_cycle = 'monthly' WHERE billing_cycle IS NULL;