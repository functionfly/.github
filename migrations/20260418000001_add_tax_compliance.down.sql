-- Tax/VAT Compliance Migration Down
-- Reverts tax-related fields and tables

-- Drop tax ID validation logs
DROP TABLE IF EXISTS tax_id_validation_logs;

-- Drop invoice tax details
DROP TABLE IF EXISTS invoice_tax_details;

-- Drop tax rates table
DROP TABLE IF EXISTS tax_rates;

-- Remove tax columns from subscriptions
ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS tax_calculation_enabled,
    DROP COLUMN IF EXISTS tax_behavior;

-- Remove tax columns from tenants
ALTER TABLE tenants
    DROP COLUMN IF EXISTS billing_country,
    DROP COLUMN IF EXISTS billing_state,
    DROP COLUMN IF EXISTS billing_postal_code,
    DROP COLUMN IF EXISTS tax_id,
    DROP COLUMN IF EXISTS tax_id_type,
    DROP COLUMN IF EXISTS tax_status,
    DROP COLUMN IF EXISTS tax_exempt,
    DROP COLUMN IF EXISTS stripe_tax_location_id,
    DROP COLUMN IF EXISTS stripe_customer_tax_id;
