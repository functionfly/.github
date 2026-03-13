DROP INDEX IF EXISTS idx_subscriptions_stripe_subscription_id;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS stripe_subscription_id;
