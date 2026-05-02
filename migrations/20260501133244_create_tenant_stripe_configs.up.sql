-- Migration: Create tenant_stripe_configs table for isolated payment processing
-- Purpose: Store per-tenant Stripe configuration enabling payment isolation
-- Date: 20260501133244

CREATE TABLE IF NOT EXISTS tenant_stripe_configs (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                   UUID NOT NULL UNIQUE,
    stripe_customer_id          TEXT NOT NULL DEFAULT '',
    stripe_connect_account_id   TEXT,
    isolated_payment_enabled   BOOLEAN NOT NULL DEFAULT false,
    payment_mode                TEXT NOT NULL DEFAULT 'platform' CHECK (payment_mode IN ('platform', 'isolated', 'connect')),
    allowed_payment_methods     TEXT NOT NULL DEFAULT '["card"]',
    default_payment_method      TEXT,
    billing_address_required    BOOLEAN NOT NULL DEFAULT true,
    tax_calculation_mode        TEXT NOT NULL DEFAULT 'automatic' CHECK (tax_calculation_mode IN ('automatic', 'manual')),
    metadata                    JSONB DEFAULT '{}',
    created_at                  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for looking up tenant by Stripe customer ID
CREATE INDEX IF NOT EXISTS idx_tenant_stripe_configs_customer_id
    ON tenant_stripe_configs(stripe_customer_id)
    WHERE stripe_customer_id != '';

-- Index for finding tenants with isolated payments enabled
CREATE INDEX IF NOT EXISTS idx_tenant_stripe_configs_isolated
    ON tenant_stripe_configs(tenant_id)
    WHERE isolated_payment_enabled = true;

-- Index for payment mode queries
CREATE INDEX IF NOT EXISTS idx_tenant_stripe_configs_payment_mode
    ON tenant_stripe_configs(payment_mode)
    WHERE payment_mode != 'platform';

-- Comment on table
COMMENT ON TABLE tenant_stripe_configs IS 'Per-tenant Stripe configuration for isolated payment processing. Each tenant can have their own Stripe customer and optional Connect account.';

-- Comments on columns
COMMENT ON COLUMN tenant_stripe_configs.id IS 'Primary key';
COMMENT ON COLUMN tenant_stripe_configs.tenant_id IS 'Reference to tenant';
COMMENT ON COLUMN tenant_stripe_configs.stripe_customer_id IS 'Tenant Stripe Customer ID (cus_xxx)';
COMMENT ON COLUMN tenant_stripe_configs.stripe_connect_account_id IS 'Stripe Connect account ID for marketplace (acct_xxx)';
COMMENT ON COLUMN tenant_stripe_configs.isolated_payment_enabled IS 'Whether tenant uses isolated payment flow';
COMMENT ON COLUMN tenant_stripe_configs.payment_mode IS 'Payment mode: platform (default), isolated (tenant isolated), connect (Stripe Connect)';
COMMENT ON COLUMN tenant_stripe_configs.allowed_payment_methods IS 'JSON array of allowed payment method types';
COMMENT ON COLUMN tenant_stripe_configs.default_payment_method IS 'Default payment method ID for this tenant';
COMMENT ON COLUMN tenant_stripe_configs.billing_address_required IS 'Whether billing address is required at checkout';
COMMENT ON COLUMN tenant_stripe_configs.tax_calculation_mode IS 'Tax calculation mode: automatic (Stripe Tax) or manual';
COMMENT ON COLUMN tenant_stripe_configs.metadata IS 'Additional tenant-specific Stripe configuration';

-- Trigger to auto-update updated_at
CREATE TRIGGER trg_tenant_stripe_configs_updated_at
    BEFORE UPDATE ON tenant_stripe_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add foreign key constraint to tenants table (if table exists)
-- This is informational - the actual constraint may already exist or need separate migration
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_name = 'tenants'
    ) THEN
        -- Note: We cannot add the FK here because tenant_stripe_configs might be created
        -- before or after tenants table depending on migration order.
        -- The application code should ensure referential integrity.
        RAISE NOTICE 'Tenants table exists - ensure referential integrity at application level';
    END IF;
END $$;

-- Insert default configs for existing tenants that don't have one
-- This is optional and can be removed if not needed
INSERT INTO tenant_stripe_configs (id, tenant_id, stripe_customer_id, isolated_payment_enabled, payment_mode, allowed_payment_methods, billing_address_required, tax_calculation_mode, metadata, created_at, updated_at)
SELECT
    gen_random_uuid(),
    t.id,
    COALESCE(t.stripe_customer_id, ''),
    false,
    'platform',
    '["card"]',
    true,
    'automatic',
    '{"migrated_from_tenants": true}'::jsonb,
    NOW(),
    NOW()
FROM tenants t
WHERE NOT EXISTS (
    SELECT 1 FROM tenant_stripe_configs tsc WHERE tsc.tenant_id = t.id
)
ON CONFLICT (tenant_id) DO NOTHING;