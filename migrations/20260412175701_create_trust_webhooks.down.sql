-- Migration: Drop trust webhook tables
-- Reverses: 20260412175701_create_trust_webhooks

-- Drop triggers first
DROP TRIGGER IF EXISTS update_trust_webhooks_updated_at ON trust_webhooks;
DROP TRIGGER IF EXISTS update_trust_webhook_deliveries_updated_at ON trust_webhook_deliveries;

-- Drop helper function
DROP FUNCTION IF EXISTS update_trust_webhook_updated_at_column();

-- Drop tables (deliveries depends on webhooks)
DROP TABLE IF EXISTS trust_webhook_deliveries;
DROP TABLE IF EXISTS trust_webhooks;