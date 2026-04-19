-- Create payment_disputes table for tracking Stripe chargebacks and disputes
CREATE TABLE IF NOT EXISTS payment_disputes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NULL,
    user_id UUID NULL,
    stripe_dispute_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_payment_id VARCHAR(255) NOT NULL,
    stripe_charge_id VARCHAR(255) NULL,
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    reason VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    evidence_due_by TIMESTAMP NULL,
    evidence_submitted BOOLEAN DEFAULT FALSE,
    evidence_data JSONB DEFAULT '{}',
    outcome VARCHAR(50) NULL,
    outcome_reason VARCHAR(255) NULL,
    network_reason_code VARCHAR(50) NULL,
    is_charge_refundable BOOLEAN DEFAULT FALSE,
    refund_id VARCHAR(255) NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    resolved_at TIMESTAMP NULL
);

-- Create indexes for payment_disputes
CREATE INDEX IF NOT EXISTS idx_payment_disputes_tenant_id ON payment_disputes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_disputes_user_id ON payment_disputes(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_disputes_stripe_payment_id ON payment_disputes(stripe_payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_disputes_stripe_charge_id ON payment_disputes(stripe_charge_id);
CREATE INDEX IF NOT EXISTS idx_payment_disputes_status ON payment_disputes(status);
CREATE INDEX IF NOT EXISTS idx_payment_disputes_created_at ON payment_disputes(created_at);

-- Create payment_refunds table for tracking Stripe refunds
CREATE TABLE IF NOT EXISTS payment_refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NULL,
    user_id UUID NULL,
    stripe_refund_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_payment_id VARCHAR(255) NOT NULL,
    stripe_charge_id VARCHAR(255) NULL,
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) NOT NULL,
    reason VARCHAR(50) NULL,
    receipt_number VARCHAR(100) NULL,
    receipt_url VARCHAR(500) NULL,
    description VARCHAR(500) NULL,
    metadata JSONB DEFAULT '{}',
    failure_reason VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- Create indexes for payment_refunds
CREATE INDEX IF NOT EXISTS idx_payment_refunds_tenant_id ON payment_refunds(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_refunds_user_id ON payment_refunds(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_refunds_stripe_payment_id ON payment_refunds(stripe_payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_refunds_stripe_charge_id ON payment_refunds(stripe_charge_id);
CREATE INDEX IF NOT EXISTS idx_payment_refunds_status ON payment_refunds(status);
CREATE INDEX IF NOT EXISTS idx_payment_refunds_created_at ON payment_refunds(created_at);

-- Add foreign key constraints (optional - can be enabled if you want strict referential integrity)
-- ALTER TABLE payment_disputes ADD CONSTRAINT fk_payment_disputes_tenant_id
--     FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL;
-- ALTER TABLE payment_disputes ADD CONSTRAINT fk_payment_disputes_user_id
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
-- ALTER TABLE payment_refunds ADD CONSTRAINT fk_payment_refunds_tenant_id
--     FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL;
-- ALTER TABLE payment_refunds ADD CONSTRAINT fk_payment_refunds_user_id
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
