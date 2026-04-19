-- Table to store tenant payment method information from Stripe
-- Enables tracking of payment method changes from Stripe dashboard
CREATE TABLE IF NOT EXISTS tenant_payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    stripe_payment_method_id VARCHAR(255) NOT NULL,
    brand VARCHAR(20), -- 'visa', 'mastercard', 'amex', etc.
    last4 VARCHAR(4),
    exp_month INTEGER,
    exp_year INTEGER,
    is_default BOOLEAN DEFAULT false,
    billing_details JSONB, -- Name, email, address from Stripe
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, stripe_payment_method_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_tenant_payment_methods_tenant_id ON tenant_payment_methods(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_payment_methods_stripe_id ON tenant_payment_methods(stripe_payment_method_id);
CREATE INDEX IF NOT EXISTS idx_tenant_payment_methods_is_default ON tenant_payment_methods(is_default);

-- Comments
COMMENT ON TABLE tenant_payment_methods IS 'Cached payment method information from Stripe for two-way sync';
COMMENT ON COLUMN tenant_payment_methods.stripe_payment_method_id IS 'Stripe PaymentMethod ID (pm_xxx)';
