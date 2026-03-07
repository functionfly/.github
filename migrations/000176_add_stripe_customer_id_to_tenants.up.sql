-- Add Stripe customer ID to tenants for billing portal and subscriptions
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS stripe_customer_id TEXT;

COMMENT ON COLUMN tenants.stripe_customer_id IS 'Stripe Customer ID for billing portal and subscription management';
