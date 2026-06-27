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
    auto_refund_threshold_cents INTEGER DEFAULT 2500, -- $25
    auto_refund_allowed_reasons TEXT[] DEFAULT ARRAY['duplicate', 'product_not_received'],
    evidence_auto_submit BOOLEAN DEFAULT false, -- Manual review by default
    evidence_auto_submit_threshold_cents INTEGER DEFAULT 15000, -- $150
    manual_review_threshold_cents INTEGER DEFAULT 15000, -- $150
    customer_notification_enabled BOOLEAN DEFAULT true,
    admin_escalation_enabled BOOLEAN DEFAULT true,
    admin_escalation_threshold_cents INTEGER DEFAULT 15000, -- $150
    fraud_detection_enabled BOOLEAN DEFAULT true,
    repeat_offender_window_days INTEGER DEFAULT 180,
    repeat_offender_threshold INTEGER DEFAULT 2,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
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

-- Table: dispute_customer_notifications
-- Tracks customer notifications sent for disputes
CREATE TABLE IF NOT EXISTS dispute_customer_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES payment_disputes(id) ON DELETE CASCADE,
    notification_type VARCHAR(50) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    content JSONB,
    success BOOLEAN DEFAULT true,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_dispute_customer_notifications_dispute_id ON dispute_customer_notifications(dispute_id);
CREATE INDEX IF NOT EXISTS idx_dispute_customer_notifications_sent_at ON dispute_customer_notifications(sent_at);

-- Additional indexes for payment_disputes
CREATE INDEX IF NOT EXISTS idx_payment_disputes_evidence_due_by ON payment_disputes(evidence_due_by) WHERE evidence_due_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_disputes_status_pending ON payment_disputes(status) WHERE status = 'pending_review';
