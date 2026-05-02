-- Migration: Drop tenant_stripe_configs table
-- Purpose: Remove isolated payment configuration table
-- Date: 20260501133244

-- Remove trigger first
DROP TRIGGER IF EXISTS trg_tenant_stripe_configs_updated_at ON tenant_stripe_configs;

-- Drop indexes
DROP INDEX IF EXISTS idx_tenant_stripe_configs_customer_id;
DROP INDEX IF EXISTS idx_tenant_stripe_configs_isolated;
DROP INDEX IF EXISTS idx_tenant_stripe_configs_payment_mode;

-- Drop table
DROP TABLE IF EXISTS tenant_stripe_configs;