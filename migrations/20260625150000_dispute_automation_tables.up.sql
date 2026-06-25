-- Chargeback Automated Response Tables

-- Table: dispute_automation_log
-- Tracks all automated decisions and actions taken for disputes
CREATE TABLE IF NOT EXISTS dispute_automation_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES payment_disputes(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    outcome VARCHAR(50) NOT NULL,
    details JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dispute_automation_log_dispute_id ON dispute_automation_log(dispute_id);
CREATE INDEX IF NOT EXISTS idx_dispute_automation_log_action ON dispute_automation_log(action);
CREATE INDEX IF NOT EXISTS idx_dispute_automation_log_created_at ON dispute_automation_log(created_at);

-- Table: dispute_automation_config
-- Configurable settings for automation behavior
CREATE TABLE IF NOT EXISTS dispute_automation_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auto_refund_enabled BOOLEAN DEFAULT true,
    auto_refund_threshold_cents INTEGER DEFAULT 5000,
    auto_refund_allowed_reasons TEXT[] DEFAULT ARRAY['duplicate', 'product_not_received'],
    evidence_auto_submit BOOLEAN DEFAULT true,
    customer_notification_enabled BOOLEAN DEFAULT true,
    admin_escalation_enabled BOOLEAN DEFAULT true,
    admin_escalation_threshold_cents INTEGER DEFAULT 50000,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default config
INSERT INTO dispute_automation_config (id) VALUES (gen_random_uuid()) ON CONFLICT DO NOTHING;

-- Table: dispute_evidence_cache
-- Caches compiled evidence to avoid re-computing
CREATE TABLE IF NOT EXISTS dispute_evidence_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES payment_disputes(id) ON DELETE CASCADE,
    evidence_data JSONB NOT NULL,
    compiled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(dispute_id)
);

CREATE INDEX IF NOT EXISTS idx_dispute_evidence_cache_dispute_id ON dispute_evidence_cache(dispute_id);
