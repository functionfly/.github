-- Table to track Stripe webhook sync events for audit and data consistency
-- This enables two-way sync between Stripe and internal records
CREATE TABLE IF NOT EXISTS stripe_sync_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stripe_event_id VARCHAR(255) NOT NULL,
    stripe_object_id VARCHAR(255) NOT NULL, -- ID of the Stripe object (subscription, payment_method, etc.)
    event_type VARCHAR(100) NOT NULL, -- e.g., 'customer.subscription.updated', 'payment_method.updated'
    event_data JSONB NOT NULL, -- Full Stripe event payload
    tenant_id UUID REFERENCES tenants(id), -- May be null if we can't find the tenant immediately
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'processed', 'failed', 'ignored'
    error_message TEXT, -- Error details if processing failed
    processed_at TIMESTAMP WITH TIME ZONE,
    retry_count INTEGER DEFAULT 0,
    idempotency_key VARCHAR(255), -- Stripe's idempotency key to prevent duplicate processing
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_stripe_sync_events_event_id ON stripe_sync_events(stripe_event_id);
CREATE INDEX IF NOT EXISTS idx_stripe_sync_events_object_id ON stripe_sync_events(stripe_object_id);
CREATE INDEX IF NOT EXISTS idx_stripe_sync_events_event_type ON stripe_sync_events(event_type);
CREATE INDEX IF NOT EXISTS idx_stripe_sync_events_tenant_id ON stripe_sync_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stripe_sync_events_status ON stripe_sync_events(status);
CREATE INDEX IF NOT EXISTS idx_stripe_sync_events_created_at ON stripe_sync_events(created_at);
CREATE INDEX IF NOT EXISTS idx_stripe_sync_events_idempotency ON stripe_sync_events(idempotency_key) WHERE idempotency_key IS NOT NULL;

-- Comments
COMMENT ON TABLE stripe_sync_events IS 'Audit trail for Stripe webhook events enabling two-way sync';
COMMENT ON COLUMN stripe_sync_events.stripe_event_id IS 'Stripe unique event ID (evt_xxx)';
COMMENT ON COLUMN stripe_sync_events.stripe_object_id IS 'ID of the object the event relates to (sub_xxx, pm_xxx, etc.)';
COMMENT ON COLUMN stripe_sync_events.event_type IS 'Stripe event type string';
COMMENT ON COLUMN stripe_sync_events.event_data IS 'Full JSON payload from Stripe';
COMMENT ON COLUMN stripe_sync_events.status IS 'Processing status of the event';
