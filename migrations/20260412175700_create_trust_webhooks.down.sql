-- Migration: Drop webhook tables
-- Description: Removes webhook infrastructure tables

-- Drop triggers
DROP TRIGGER IF EXISTS update_trust_webhooks_updated_at ON trust_webhooks;
DROP TRIGGER IF EXISTS update_trust_webhook_deliveries_updated_at ON trust_webhook_deliveries;

-- Drop function
DROP FUNCTION IF EXISTS update_trust_webhook_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS trust_webhook_deliveries;
DROP TABLE IF EXISTS trust_webhooks;
