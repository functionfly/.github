-- Migration: Create Stripe usage reporting tracking table for metered billing
-- Created: 2026-04-19

-- Table: stripe_usage_reports
-- Tracks usage reported to Stripe to prevent double-billing and ensure accurate metered billing
CREATE TABLE IF NOT EXISTS stripe_usage_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- References
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    partner_id UUID, -- For Trust API partners (optional, for non-trust usage)
    subscription_id VARCHAR(255) NOT NULL, -- Stripe subscription ID
    subscription_item_id VARCHAR(255) NOT NULL, -- Stripe subscription item ID (metered price)
    
    -- Usage being reported
    usage_quantity INTEGER NOT NULL, -- Number of units reported
    usage_period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    usage_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Stripe API tracking
    stripe_timestamp INTEGER NOT NULL, -- Unix timestamp when reported to Stripe
    stripe_usage_record_id VARCHAR(255), -- ID returned by Stripe (if available)
    
    -- Reporting status
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'reported', 'failed', 'reconciled'
    
    -- Error tracking (for failed reports)
    error_message TEXT,
    retry_count INTEGER NOT NULL DEFAULT 0,
    
    -- Idempotency key for Stripe API
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    
    -- Metadata
    meter_event_name VARCHAR(255), -- e.g., 'trust_api_overage', 'function_execution'
    metadata JSONB DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for stripe_usage_reports
CREATE INDEX IF NOT EXISTS idx_stripe_usage_reports_tenant_id ON stripe_usage_reports(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stripe_usage_reports_partner_id ON stripe_usage_reports(partner_id) WHERE partner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_stripe_usage_reports_subscription_id ON stripe_usage_reports(subscription_id);
CREATE INDEX IF NOT EXISTS idx_stripe_usage_reports_status ON stripe_usage_reports(status);
CREATE INDEX IF NOT EXISTS idx_stripe_usage_reports_period ON stripe_usage_reports(usage_period_start, usage_period_end);
CREATE INDEX IF NOT EXISTS idx_stripe_usage_reports_created_at ON stripe_usage_reports(created_at);

-- Unique constraint to prevent duplicate reporting for same period
CREATE UNIQUE INDEX IF NOT EXISTS idx_stripe_usage_reports_unique_period 
ON stripe_usage_reports(tenant_id, subscription_item_id, usage_period_start, usage_period_end) 
WHERE status IN ('reported', 'reconciled');

-- Table: billing_usage_reconciliation
-- Tracks reconciliation between internal usage tracking and Stripe reported usage
CREATE TABLE IF NOT EXISTS billing_usage_reconciliation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subscription_id VARCHAR(255) NOT NULL,
    
    -- Period
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Internal tracking
    internal_usage_count INTEGER NOT NULL,
    internal_usage_value INTEGER NOT NULL, -- In cents or units
    
    -- Stripe reported
    stripe_reported_count INTEGER,
    stripe_reported_value INTEGER,
    
    -- Reconciliation status
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'matched', 'discrepancy', 'resolved'
    discrepancy_amount INTEGER, -- If there's a difference
    discrepancy_reason TEXT,
    
    -- Resolution
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    resolution_notes TEXT,
    
    metadata JSONB DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Ensure one reconciliation record per period per subscription
    CONSTRAINT unique_reconciliation_period UNIQUE (tenant_id, subscription_id, period_start, period_end)
);

CREATE INDEX IF NOT EXISTS idx_billing_usage_reconciliation_tenant ON billing_usage_reconciliation(tenant_id);
CREATE INDEX IF NOT EXISTS idx_billing_usage_reconciliation_status ON billing_usage_reconciliation(status) WHERE status != 'matched';

-- Trigger function for updated_at
CREATE OR REPLACE FUNCTION update_stripe_usage_report_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing triggers if they exist
DROP TRIGGER IF EXISTS trg_stripe_usage_reports_updated_at ON stripe_usage_reports;
DROP TRIGGER IF EXISTS trg_billing_usage_reconciliation_updated_at ON billing_usage_reconciliation;

-- Create triggers
CREATE TRIGGER trg_stripe_usage_reports_updated_at
    BEFORE UPDATE ON stripe_usage_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_stripe_usage_report_updated_at();

CREATE TRIGGER trg_billing_usage_reconciliation_updated_at
    BEFORE UPDATE ON billing_usage_reconciliation
    FOR EACH ROW
    EXECUTE FUNCTION update_stripe_usage_report_updated_at();

-- Add comments
COMMENT ON TABLE stripe_usage_reports IS 'Tracks usage reported to Stripe for metered billing to prevent double-billing';
COMMENT ON TABLE billing_usage_reconciliation IS 'Reconciliation records between internal usage and Stripe reported usage';
