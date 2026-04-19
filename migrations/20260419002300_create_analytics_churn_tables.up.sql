-- Migration: Create subscription churn events table for MRR/ARR and churn analytics
-- This enables tracking of cancellations, downgrades, and failed renewals for analytics

CREATE TABLE IF NOT EXISTS subscription_churn_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    subscription_id UUID NOT NULL,
    bundle_id UUID,
    
    -- Event classification
    event_type VARCHAR(50) NOT NULL, -- 'cancellation', 'downgrade', 'upgrade', 'failed_renewal', 'payment_failure'
    event_date TIMESTAMP WITH TIME ZONE NOT NULL,
    cohort_month VARCHAR(7), -- YYYY-MM format for cohort analysis
    
    -- Financial impact
    previous_mrr_cents INTEGER DEFAULT 0,
    new_mrr_cents INTEGER DEFAULT 0,
    mrr_lost_cents INTEGER DEFAULT 0,
    mrr_gained_cents INTEGER DEFAULT 0,
    
    -- Downgrade specific
    previous_bundle_id UUID,
    new_bundle_id UUID,
    
    -- Cancellation specific
    cancel_reason VARCHAR(255),
    cancel_at_period_end BOOLEAN DEFAULT false,
    cancel_date TIMESTAMP WITH TIME ZONE,
    
    -- Payment failure specific
    failed_payment_amount_cents INTEGER DEFAULT 0,
    failure_code VARCHAR(100),
    attempt_count INTEGER DEFAULT 0,
    
    -- Stripe reference
    stripe_subscription_id VARCHAR(255),
    stripe_event_id VARCHAR(255),
    
    -- Recovery tracking
    is_recovered BOOLEAN DEFAULT false,
    recovered_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_churn_events_tenant ON subscription_churn_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_churn_events_subscription ON subscription_churn_events(subscription_id);
CREATE INDEX IF NOT EXISTS idx_churn_events_type ON subscription_churn_events(event_type);
CREATE INDEX IF NOT EXISTS idx_churn_events_date ON subscription_churn_events(event_date);
CREATE INDEX IF NOT EXISTS idx_churn_events_cohort ON subscription_churn_events(cohort_month);
CREATE INDEX IF NOT EXISTS idx_churn_events_stripe ON subscription_churn_events(stripe_subscription_id);

-- Composite index for analytics queries
CREATE INDEX IF NOT EXISTS idx_churn_events_analytics ON subscription_churn_events(event_type, event_date, is_recovered);

-- Add trigger to update updated_at
CREATE OR REPLACE FUNCTION update_churn_event_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_churn_event_updated_at ON subscription_churn_events;
CREATE TRIGGER trigger_update_churn_event_updated_at
    BEFORE UPDATE ON subscription_churn_events
    FOR EACH ROW
    EXECUTE FUNCTION update_churn_event_updated_at();

-- Add comments
COMMENT ON TABLE subscription_churn_events IS 'Tracks subscription lifecycle events for MRR/ARR and churn analytics';
COMMENT ON COLUMN subscription_churn_events.event_type IS 'Type of event: cancellation, downgrade, upgrade, failed_renewal, payment_failure';
COMMENT ON COLUMN subscription_churn_events.cohort_month IS 'YYYY-MM format for cohort retention analysis';
COMMENT ON COLUMN subscription_churn_events.mrr_lost_cents IS 'Amount of MRR lost due to this event';
COMMENT ON COLUMN subscription_churn_events.is_recovered IS 'True if customer reactivated after churn';

-- Revenue recognition table for accrual accounting support
CREATE TABLE IF NOT EXISTS revenue_recognition (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    
    -- Recognition period
    recognition_month VARCHAR(7) NOT NULL, -- YYYY-MM
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    
    -- Amounts
    total_invoice_cents INTEGER NOT NULL DEFAULT 0,
    recognized_cents INTEGER NOT NULL DEFAULT 0,
    deferred_cents INTEGER NOT NULL DEFAULT 0,
    
    -- Revenue type
    revenue_type VARCHAR(50) NOT NULL, -- 'subscription', 'usage', 'one_time'
    
    -- Status
    is_recognized BOOLEAN DEFAULT false,
    recognized_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for revenue recognition
CREATE INDEX IF NOT EXISTS idx_revenue_recognition_invoice ON revenue_recognition(invoice_id);
CREATE INDEX IF NOT EXISTS idx_revenue_recognition_tenant ON revenue_recognition(tenant_id);
CREATE INDEX IF NOT EXISTS idx_revenue_recognition_month ON revenue_recognition(recognition_month);
CREATE INDEX IF NOT EXISTS idx_revenue_recognition_type ON revenue_recognition(revenue_type);

-- Unique constraint to prevent duplicate recognition entries
CREATE UNIQUE INDEX IF NOT EXISTS idx_revenue_recognition_unique ON revenue_recognition(invoice_id, recognition_month);

COMMENT ON TABLE revenue_recognition IS 'Tracks revenue recognition for accrual accounting';
COMMENT ON COLUMN revenue_recognition.recognized_cents IS 'Amount of revenue recognized in this period';
COMMENT ON COLUMN revenue_recognition.deferred_cents IS 'Amount of revenue deferred to future periods';

-- MRR snapshot table for historical tracking
CREATE TABLE IF NOT EXISTS mrr_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_date DATE NOT NULL,
    
    -- MRR breakdown
    total_mrr_cents INTEGER NOT NULL DEFAULT 0,
    new_mrr_cents INTEGER NOT NULL DEFAULT 0,
    expansion_mrr_cents INTEGER NOT NULL DEFAULT 0,
    contraction_mrr_cents INTEGER NOT NULL DEFAULT 0,
    churned_mrr_cents INTEGER NOT NULL DEFAULT 0,
    reactivation_mrr_cents INTEGER NOT NULL DEFAULT 0,
    
    -- Customer counts
    active_customers INTEGER NOT NULL DEFAULT 0,
    new_customers INTEGER NOT NULL DEFAULT 0,
    churned_customers INTEGER NOT NULL DEFAULT 0,
    
    -- Period
    period_month VARCHAR(7) NOT NULL, -- YYYY-MM
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Unique constraint for daily snapshots
CREATE UNIQUE INDEX IF NOT EXISTS idx_mrr_snapshots_date ON mrr_snapshots(snapshot_date);
CREATE INDEX IF NOT EXISTS idx_mrr_snapshots_month ON mrr_snapshots(period_month);

COMMENT ON TABLE mrr_snapshots IS 'Daily snapshots of MRR for historical trend analysis';

-- Cohort analysis table
CREATE TABLE IF NOT EXISTS cohort_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cohort_month VARCHAR(7) NOT NULL, -- YYYY-MM
    
    -- Cohort definition
    cohort_size INTEGER NOT NULL DEFAULT 0,
    
    -- Period tracking
    period_index INTEGER NOT NULL DEFAULT 0, -- 0 = first month, 1 = second month, etc.
    period_date DATE NOT NULL,
    
    -- Retention metrics
    active_customers INTEGER NOT NULL DEFAULT 0,
    retention_rate DECIMAL(5,2), -- percentage
    
    -- Revenue
    revenue_cents INTEGER NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(cohort_month, period_index)
);

CREATE INDEX IF NOT EXISTS idx_cohort_analysis_month ON cohort_analysis(cohort_month);
CREATE INDEX IF NOT EXISTS idx_cohort_analysis_period ON cohort_analysis(period_date);

COMMENT ON TABLE cohort_analysis IS 'Cohort retention analysis data';

-- Failed payment tracking table for dunning analysis
CREATE TABLE IF NOT EXISTS failed_payment_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    invoice_id UUID,
    
    -- Payment details
    failed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Failure details
    failure_code VARCHAR(100),
    failure_message TEXT,
    attempt_number INTEGER DEFAULT 1,
    
    -- Recovery tracking
    recovered BOOLEAN DEFAULT false,
    recovered_at TIMESTAMP WITH TIME ZONE,
    recovery_method VARCHAR(50), -- 'retry', 'customer_action', 'dunning'
    
    -- Churn outcome
    led_to_churn BOOLEAN DEFAULT false,
    
    -- Stripe reference
    stripe_payment_intent_id VARCHAR(255),
    stripe_invoice_id VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_failed_payments_tenant ON failed_payment_analytics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_failed_payments_failed_at ON failed_payment_analytics(failed_at);
CREATE INDEX IF NOT EXISTS idx_failed_payments_recovered ON failed_payment_analytics(recovered);
CREATE INDEX IF NOT EXISTS idx_failed_payments_churn ON failed_payment_analytics(led_to_churn, recovered);

COMMENT ON TABLE failed_payment_analytics IS 'Tracks failed payments for dunning and churn analysis';
COMMENT ON COLUMN failed_payment_analytics.recovery_method IS 'How the payment was recovered: retry, customer_action, dunning';
COMMENT ON COLUMN failed_payment_analytics.led_to_churn IS 'Whether this failure eventually led to customer churn';

-- Trigger for failed payment analytics updated_at
DROP TRIGGER IF EXISTS trigger_update_failed_payment_updated_at ON failed_payment_analytics;
CREATE TRIGGER trigger_update_failed_payment_updated_at
    BEFORE UPDATE ON failed_payment_analytics
    FOR EACH ROW
    EXECUTE FUNCTION update_churn_event_updated_at();

-- Add GIN index for any JSON metadata fields if needed
-- Note: If we add metadata JSONB columns later, use GIN indexes

-- Grant permissions (to be run manually by admin)
-- GRANT SELECT, INSERT, UPDATE ON subscription_churn_events TO app_role;
-- GRANT SELECT, INSERT, UPDATE ON revenue_recognition TO app_role;
-- GRANT SELECT, INSERT ON mrr_snapshots TO app_role;
-- GRANT SELECT, INSERT ON cohort_analysis TO app_role;
-- GRANT SELECT, INSERT, UPDATE ON failed_payment_analytics TO app_role;