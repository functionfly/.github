-- Tax/VAT Compliance Migration
-- Adds tax-related fields for Stripe Tax integration
-- Supports EU VAT, US sales tax, and global tax compliance

-- Add tax fields to tenants table
ALTER TABLE tenants 
    ADD COLUMN IF NOT EXISTS billing_country VARCHAR(2),
    ADD COLUMN IF NOT EXISTS billing_state VARCHAR(50),
    ADD COLUMN IF NOT EXISTS billing_postal_code VARCHAR(20),
    ADD COLUMN IF NOT EXISTS tax_id VARCHAR(50),
    ADD COLUMN IF NOT EXISTS tax_id_type VARCHAR(20),
    ADD COLUMN IF NOT EXISTS tax_status VARCHAR(20) DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS tax_exempt BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS stripe_tax_location_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS stripe_customer_tax_id VARCHAR(255);

-- Create tax rates table for caching Stripe tax rates
CREATE TABLE IF NOT EXISTS tax_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country VARCHAR(2) NOT NULL,
    state VARCHAR(50),
    postal_code VARCHAR(20),
    tax_type VARCHAR(50) NOT NULL, -- 'vat', 'sales_tax', 'gst', etc.
    display_name VARCHAR(100) NOT NULL,
    percentage DECIMAL(5,2) NOT NULL,
    inclusive BOOLEAN DEFAULT FALSE,
    stripe_tax_rate_id VARCHAR(255),
    jurisdiction VARCHAR(100),
    effective_from TIMESTAMP WITH TIME ZONE,
    effective_until TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create invoice tax details table for storing tax breakdown per invoice
CREATE TABLE IF NOT EXISTS invoice_tax_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    tax_rate_id UUID REFERENCES tax_rates(id),
    tax_amount_cents INTEGER NOT NULL DEFAULT 0,
    subtotal_cents INTEGER NOT NULL DEFAULT 0,
    total_cents INTEGER NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    tax_name VARCHAR(100),
    tax_percentage DECIMAL(5,2),
    stripe_tax_calculation_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create tax ID validation log for audit purposes
CREATE TABLE IF NOT EXISTS tax_id_validation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    tax_id VARCHAR(50) NOT NULL,
    tax_id_type VARCHAR(20) NOT NULL,
    validation_status VARCHAR(20) NOT NULL, -- 'valid', 'invalid', 'pending'
    validation_source VARCHAR(50), -- 'vies', 'stripe', 'manual'
    validation_response TEXT,
    validated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for tax rate lookups
CREATE INDEX IF NOT EXISTS idx_tax_rates_country_state ON tax_rates(country, state);
CREATE INDEX IF NOT EXISTS idx_tax_rates_postal ON tax_rates(postal_code);
CREATE INDEX IF NOT EXISTS idx_tax_rates_active ON tax_rates(is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_tax_rates_stripe_id ON tax_rates(stripe_tax_rate_id);

-- Create indexes for invoice tax details
CREATE INDEX IF NOT EXISTS idx_invoice_tax_invoice ON invoice_tax_details(invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoice_tax_rate ON invoice_tax_details(tax_rate_id);

-- Create indexes for tax ID validation logs
CREATE INDEX IF NOT EXISTS idx_tax_validation_tenant ON tax_id_validation_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tax_validation_status ON tax_id_validation_logs(validation_status);

-- Create index for tenant tax lookups
CREATE INDEX IF NOT EXISTS idx_tenants_billing_country ON tenants(billing_country);
CREATE INDEX IF NOT EXISTS idx_tenants_tax_status ON tenants(tax_status);

-- Add comments for documentation
COMMENT ON COLUMN tenants.billing_country IS 'ISO 3166-1 alpha-2 country code for billing/tax purposes';
COMMENT ON COLUMN tenants.billing_state IS 'State/Province for tax calculation (required for US, Canada, etc.)';
COMMENT ON COLUMN tenants.billing_postal_code IS 'Postal/ZIP code for tax jurisdiction determination';
COMMENT ON COLUMN tenants.tax_id IS 'Tax ID number (VAT number, EIN, etc.)';
COMMENT ON COLUMN tenants.tax_id_type IS 'Type of tax ID: eu_vat, us_ein, ca_gst, uk_vat, etc.';
COMMENT ON COLUMN tenants.tax_status IS 'Tax status: pending, valid, invalid, exempt';
COMMENT ON COLUMN tenants.tax_exempt IS 'Whether the customer is tax exempt';
COMMENT ON COLUMN tenants.stripe_tax_location_id IS 'Stripe Tax Location ID for automatic tax calculation';
COMMENT ON COLUMN tenants.stripe_customer_tax_id IS 'Stripe Customer Tax ID object ID';

COMMENT ON TABLE tax_rates IS 'Cached tax rates from Stripe Tax for reporting';
COMMENT ON TABLE invoice_tax_details IS 'Tax breakdown details per invoice';
COMMENT ON TABLE tax_id_validation_logs IS 'Audit log for tax ID validation attempts';

-- Add tax-related columns to subscriptions for tax calculation tracking
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS tax_calculation_enabled BOOLEAN DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS tax_behavior VARCHAR(20) DEFAULT 'exclusive'; -- 'exclusive', 'inclusive'
