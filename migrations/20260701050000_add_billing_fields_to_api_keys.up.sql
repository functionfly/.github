-- Migration: 20260701050000_add_billing_fields_to_api_keys
-- Description: Add per-key billing attribution columns to api_keys
-- Created: 2026-07-01

ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS billing_budget_cents BIGINT DEFAULT 0;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS is_high_value BOOLEAN DEFAULT false;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS cost_center VARCHAR(100);

COMMENT ON COLUMN api_keys.billing_budget_cents IS 'Optional per-key budget in cents for high-value key attribution';
COMMENT ON COLUMN api_keys.is_high_value IS 'Mark as high-value key for separate billing';
COMMENT ON COLUMN api_keys.cost_center IS 'Cost center for chargeback reporting';
