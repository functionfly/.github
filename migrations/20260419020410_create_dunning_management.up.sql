-- Migration: Create dunning management tables for automated payment retry logic
-- Created: 2026-04-19

-- Table: payment_retry_schedules
-- Defines retry schedules with configurable intervals and grace periods
CREATE TABLE IF NOT EXISTS payment_retry_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Retry configuration
    max_retries INTEGER NOT NULL DEFAULT 4,
    retry_intervals INTEGER[] NOT NULL DEFAULT ARRAY[1, 3, 7, 14], -- Days between retries (1 day, 3 days, 7 days, 14 days)
    
    -- Grace period configuration
    grace_period_days INTEGER NOT NULL DEFAULT 14, -- Service continues during this period
    
    -- Escalation settings
    send_customer_notifications BOOLEAN NOT NULL DEFAULT true,
    notify_admin_on_final_retry BOOLEAN NOT NULL DEFAULT true,
    suspend_service_after_final_retry BOOLEAN NOT NULL DEFAULT true,
    
    -- Schedule assignment
    schedule_type VARCHAR(50) NOT NULL DEFAULT 'default', -- 'default', 'enterprise', 'startup', etc.
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_retry_intervals CHECK (array_length(retry_intervals, 1) >= max_retries)
);

-- Create default retry schedule
INSERT INTO payment_retry_schedules (
    name, 
    description, 
    max_retries, 
    retry_intervals, 
    grace_period_days,
    schedule_type
) VALUES (
    'Standard Retry Schedule',
    'Default retry schedule: retries at 1, 3, 7, and 14 days with 14-day grace period',
    4,
    ARRAY[1, 3, 7, 14],
    14,
    'default'
) ON CONFLICT DO NOTHING;

-- Table: payment_retries
-- Tracks individual payment retry attempts for failed invoices
CREATE TABLE IF NOT EXISTS payment_retries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- References
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    invoice_id VARCHAR(255), -- Stripe invoice ID
    stripe_customer_id VARCHAR(255),
    
    -- Retry schedule reference
    schedule_id UUID REFERENCES payment_retry_schedules(id),
    
    -- Retry state
    current_attempt INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 4,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'paused', 'resolved', 'failed', 'cancelled'
    
    -- Financial details
    amount_due_cents INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    
    -- Timing
    initial_failure_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_retry_at TIMESTAMP WITH TIME ZONE,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    grace_period_ends_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Resolution tracking
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_type VARCHAR(50), -- 'payment_success', 'manual_payment', 'subscription_cancelled', 'written_off'
    
    -- Retry history (JSON array of attempt details)
    retry_history JSONB DEFAULT '[]'::jsonb,
    
    -- Failure reason tracking
    last_failure_code VARCHAR(100),
    last_failure_message TEXT,
    decline_code VARCHAR(100),
    
    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for payment_retries
CREATE INDEX IF NOT EXISTS idx_payment_retries_tenant_id ON payment_retries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_retries_status ON payment_retries(status) WHERE status IN ('active', 'paused');
CREATE INDEX IF NOT EXISTS idx_payment_retries_next_retry_at ON payment_retries(next_retry_at) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_payment_retries_grace_period ON payment_retries(grace_period_ends_at) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_payment_retries_invoice_id ON payment_retries(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payment_retries_subscription_id ON payment_retries(subscription_id);

-- Table: dunning_notifications
-- Tracks notifications sent during dunning process
CREATE TABLE IF NOT EXISTS dunning_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    payment_retry_id UUID NOT NULL REFERENCES payment_retries(id) ON DELETE CASCADE,
    
    -- Notification details
    notification_type VARCHAR(50) NOT NULL, -- 'initial_failure', 'retry_reminder', 'final_notice', 'service_suspension_warning'
    attempt_number INTEGER, -- Which retry attempt this notification is for (null for initial)
    
    -- Recipient
    recipient_email VARCHAR(255) NOT NULL,
    recipient_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    
    -- Content
    subject TEXT NOT NULL,
    body TEXT,
    
    -- Status
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE,
    opened_at TIMESTAMP WITH TIME ZONE,
    clicked_at TIMESTAMP WITH TIME ZONE,
    
    -- External tracking
    email_provider_message_id VARCHAR(255),
    
    metadata JSONB DEFAULT '{}'::jsonb,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dunning_notifications_retry_id ON dunning_notifications(payment_retry_id);
CREATE INDEX IF NOT EXISTS idx_dunning_notifications_recipient ON dunning_notifications(recipient_user_id);
CREATE INDEX IF NOT EXISTS idx_dunning_notifications_sent_at ON dunning_notifications(sent_at);

-- Table: service_suspensions
-- Tracks when service is suspended due to failed payments
CREATE TABLE IF NOT EXISTS service_suspensions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    payment_retry_id UUID NOT NULL REFERENCES payment_retries(id) ON DELETE CASCADE,
    
    -- Suspension details
    suspended_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    suspended_by VARCHAR(50) NOT NULL DEFAULT 'system', -- 'system', 'admin', 'manual'
    reason VARCHAR(255) NOT NULL,
    
    -- Restoration
    restored_at TIMESTAMP WITH TIME ZONE,
    restored_by VARCHAR(50),
    restoration_reason TEXT,
    
    -- Impact tracking
    suspended_features JSONB DEFAULT '[]'::jsonb, -- Array of feature names that were suspended
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_suspensions_tenant_id ON service_suspensions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_service_suspensions_active ON service_suspensions(tenant_id, suspended_at) WHERE restored_at IS NULL;

-- Trigger function to update updated_at
CREATE OR REPLACE FUNCTION update_payment_retry_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing triggers if they exist (for idempotency)
DROP TRIGGER IF EXISTS trg_payment_retry_updated_at ON payment_retries;
DROP TRIGGER IF EXISTS trg_service_suspension_updated_at ON service_suspensions;

-- Create triggers
CREATE TRIGGER trg_payment_retry_updated_at
    BEFORE UPDATE ON payment_retries
    FOR EACH ROW
    EXECUTE FUNCTION update_payment_retry_updated_at();

CREATE TRIGGER trg_service_suspension_updated_at
    BEFORE UPDATE ON service_suspensions
    FOR EACH ROW
    EXECUTE FUNCTION update_payment_retry_updated_at();

-- Add audit logging trigger for payment_retries
CREATE OR REPLACE FUNCTION audit_payment_retry_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        -- Log status changes
        IF OLD.status IS DISTINCT FROM NEW.status THEN
            INSERT INTO audit_log (
                table_name, 
                record_id, 
                action, 
                old_values, 
                new_values, 
                performed_at
            ) VALUES (
                'payment_retries',
                NEW.id,
                'status_change',
                jsonb_build_object('status', OLD.status),
                jsonb_build_object('status', NEW.status),
                NOW()
            );
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_audit_payment_retry ON payment_retries;
CREATE TRIGGER trg_audit_payment_retry
    AFTER UPDATE ON payment_retries
    FOR EACH ROW
    EXECUTE FUNCTION audit_payment_retry_changes();

-- Add comment for documentation
COMMENT ON TABLE payment_retry_schedules IS 'Configuration for payment retry schedules including retry intervals and grace periods';
COMMENT ON TABLE payment_retries IS 'Tracks individual payment retry workflows for failed invoice payments';
COMMENT ON TABLE dunning_notifications IS 'Records notifications sent during the dunning process';
COMMENT ON TABLE service_suspensions IS 'Tracks service suspensions due to failed payment recovery';
