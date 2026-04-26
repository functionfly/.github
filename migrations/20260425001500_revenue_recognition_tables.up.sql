-- Migration: Create revenue recognition tables for ASC 606/IFRS 15 compliance
-- Implements deferred revenue tracking and recognition scheduling

CREATE TABLE IF NOT EXISTS performance_obligations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    invoice_id UUID NOT NULL,

    name VARCHAR(255) NOT NULL,
    description VARCHAR(1000),
    type VARCHAR(50) NOT NULL, -- access, usage, license, support, custom

    -- Transaction price allocation
    transaction_price_cents INTEGER NOT NULL DEFAULT 0,
    allocated_price_cents INTEGER NOT NULL DEFAULT 0,

    -- SSP (Standalone Selling Price)
    ssp_cents INTEGER NOT NULL DEFAULT 0,
    ssp_currency VARCHAR(3) DEFAULT 'USD',
    ssp_basis VARCHAR(50), -- total, per_unit, tiered

    -- Recognition pattern
    recognition_method VARCHAR(50) NOT NULL, -- over_time, point_in_time
    recognition_start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    recognition_end_date TIMESTAMP WITH TIME ZONE,
    delivery_pattern VARCHAR(50), -- linear, milestone, usage_based
    milestones JSONB DEFAULT '[]', -- [{name, date, percentage}]

    -- For usage-based
    billable_period_start TIMESTAMP WITH TIME ZONE,
    billable_period_end TIMESTAMP WITH TIME ZONE,

    -- Delivery status
    is_delivered BOOLEAN DEFAULT false,
    delivered_at TIMESTAMP WITH TIME ZONE,

    -- Recognition status
    is_fully_recognized BOOLEAN DEFAULT false,
    fully_recognized_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_po_tenant ON performance_obligations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_po_invoice ON performance_obligations(invoice_id);
CREATE INDEX IF NOT EXISTS idx_po_delivery ON performance_obligations(is_delivered, is_fully_recognized);

-- Contract assets and liabilities for balance sheet tracking
CREATE TABLE IF NOT EXISTS contract_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    invoice_id UUID,
    customer_id VARCHAR(255) NOT NULL,

    asset_type VARCHAR(50) NOT NULL, -- contract_asset, contract_liability
    amount_cents INTEGER NOT NULL DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',

    description VARCHAR(500),

    reporting_period VARCHAR(7) NOT NULL, -- YYYY-MM

    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, reduced, settled
    reduced_amount_cents INTEGER DEFAULT 0,

    is_reversed BOOLEAN DEFAULT false,
    reversed_at TIMESTAMP WITH TIME ZONE,
    reduction_reason VARCHAR(255),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ca_tenant ON contract_assets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ca_period ON contract_assets(reporting_period);
CREATE INDEX IF NOT EXISTS idx_ca_invoice ON contract_assets(invoice_id);
CREATE INDEX IF NOT EXISTS idx_ca_type_period ON contract_assets(asset_type, reporting_period);

-- Revenue recognition schedules for tracking deferred revenue
CREATE TABLE IF NOT EXISTS revenue_recognition_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    invoice_id UUID NOT NULL,
    performance_obligation_id UUID,

    recognition_month VARCHAR(7) NOT NULL, -- YYYY-MM
    period_start_date DATE NOT NULL,
    period_end_date DATE NOT NULL,

    allocated_amount_cents INTEGER NOT NULL DEFAULT 0,
    recognized_amount_cents INTEGER NOT NULL DEFAULT 0,
    deferred_amount_cents INTEGER NOT NULL DEFAULT 0,

    revenue_type VARCHAR(50) NOT NULL, -- subscription, usage, one_time

    is_recognized BOOLEAN DEFAULT false,
    recognized_at TIMESTAMP WITH TIME ZONE,

    original_total_cents INTEGER DEFAULT 0,

    is_adjustment BOOLEAN DEFAULT false,
    adjustment_reason VARCHAR(255),
    previous_schedule_id UUID,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rrs_tenant ON revenue_recognition_schedules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_rrs_invoice ON revenue_recognition_schedules(invoice_id);
CREATE INDEX IF NOT EXISTS idx_rrs_period ON revenue_recognition_schedules(recognition_month);
CREATE INDEX IF NOT EXISTS idx_rrs_po ON revenue_recognition_schedules(performance_obligation_id);
CREATE INDEX IF NOT EXISTS idx_rrs_status ON revenue_recognition_schedules(is_recognized);

-- Unique constraint to prevent duplicate schedule entries per month
CREATE UNIQUE INDEX IF NOT EXISTS idx_rrs_unique ON revenue_recognition_schedules(invoice_id, recognition_month, COALESCE(performance_obligation_id, '00000000-0000-0000-0000-000000000000'::UUID));

-- Revenue recognition events for audit trail
CREATE TABLE IF NOT EXISTS revenue_recognition_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    invoice_id UUID NOT NULL,

    event_type VARCHAR(50) NOT NULL, -- invoice_paid, delivery_completed, milestone_reached, contract_modified, credit_note_issued
    revenue_type VARCHAR(50) NOT NULL, -- subscription, usage, one_time

    gross_amount_cents INTEGER NOT NULL DEFAULT 0,
    deferred_amount_cents INTEGER NOT NULL DEFAULT 0,
    recognized_amount_cents INTEGER NOT NULL DEFAULT 0,

    event_date TIMESTAMP WITH TIME ZONE NOT NULL,
    reporting_period VARCHAR(7), -- YYYY-MM

    performance_obligation_id UUID,
    schedule_id UUID,

    previous_invoice_id UUID,
    modification_type VARCHAR(50), -- add_scope, terminate, price_change

    description VARCHAR(500),
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rre_tenant ON revenue_recognition_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_rre_invoice ON revenue_recognition_events(invoice_id);
CREATE INDEX IF NOT EXISTS idx_rre_period ON revenue_recognition_events(reporting_period);
CREATE INDEX IF NOT EXISTS idx_rre_po ON revenue_recognition_events(performance_obligation_id);
CREATE INDEX IF NOT EXISTS idx_rre_schedule ON revenue_recognition_events(schedule_id);

-- Add trigger for updated_at
CREATE OR REPLACE FUNCTION update_revenue_recognition_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_po_updated_at ON performance_obligations;
CREATE TRIGGER trigger_po_updated_at
    BEFORE UPDATE ON performance_obligations
    FOR EACH ROW EXECUTE FUNCTION update_revenue_recognition_updated_at();

DROP TRIGGER IF EXISTS trigger_ca_updated_at ON contract_assets;
CREATE TRIGGER trigger_ca_updated_at
    BEFORE UPDATE ON contract_assets
    FOR EACH ROW EXECUTE FUNCTION update_revenue_recognition_updated_at();

DROP TRIGGER IF EXISTS trigger_rrs_updated_at ON revenue_recognition_schedules;
CREATE TRIGGER trigger_rrs_updated_at
    BEFORE UPDATE ON revenue_recognition_schedules
    FOR EACH ROW EXECUTE FUNCTION update_revenue_recognition_updated_at();

COMMENT ON TABLE performance_obligations IS 'Tracks performance obligations per ASC 606/IFRS 15';
COMMENT ON TABLE contract_assets IS 'Tracks contract assets and liabilities for balance sheet';
COMMENT ON TABLE revenue_recognition_schedules IS 'Tracks deferred revenue and recognition schedules';
COMMENT ON TABLE revenue_recognition_events IS 'Audit trail for all revenue recognition events';