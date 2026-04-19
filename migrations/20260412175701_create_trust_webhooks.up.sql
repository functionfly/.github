-- Migration: Create webhook tables for trust notifications
-- Description: Adds webhook infrastructure for real-time trust event notifications

-- ============================================
-- Trust Webhooks Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL,
    owner_type VARCHAR(20) NOT NULL DEFAULT 'user',
    owner_partner_id UUID,
    url VARCHAR(500) NOT NULL,
    method VARCHAR(10) NOT NULL DEFAULT 'POST',
    secret VARCHAR(255) NOT NULL,
    events JSONB NOT NULL,
    event_filter VARCHAR(50) DEFAULT 'specific',
    function_filter JSONB DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    fail_count INTEGER DEFAULT 0,
    last_failure TIMESTAMP WITH TIME ZONE,
    last_success TIMESTAMP WITH TIME ZONE,
    max_retries INTEGER DEFAULT 3,
    retry_delay_secs INTEGER DEFAULT 60,
    timeout_secs INTEGER DEFAULT 30,
    include_payload BOOLEAN DEFAULT TRUE,
    custom_headers JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for webhooks
CREATE INDEX IF NOT EXISTS idx_trust_webhooks_owner ON trust_webhooks(owner_id, owner_type);
CREATE INDEX IF NOT EXISTS idx_trust_webhooks_status ON trust_webhooks(status);
CREATE INDEX IF NOT EXISTS idx_trust_webhooks_webhook_id ON trust_webhooks(webhook_id);

-- ============================================
-- Trust Webhook Deliveries Table
-- ============================================
CREATE TABLE IF NOT EXISTS trust_webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_id VARCHAR(32) NOT NULL UNIQUE,
    webhook_id UUID NOT NULL REFERENCES trust_webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(255),
    payload JSONB NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    attempt_number INTEGER DEFAULT 1,
    max_attempts INTEGER DEFAULT 3,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    response_status_code INTEGER,
    response_headers JSONB,
    response_body TEXT,
    response_time_ms INTEGER,
    error_message TEXT,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for deliveries
CREATE INDEX IF NOT EXISTS idx_trust_deliveries_webhook ON trust_webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_trust_deliveries_status ON trust_webhook_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_trust_deliveries_delivery_id ON trust_webhook_deliveries(delivery_id);
CREATE INDEX IF NOT EXISTS idx_trust_deliveries_event_type ON trust_webhook_deliveries(event_type);
CREATE INDEX IF NOT EXISTS idx_trust_deliveries_created_at ON trust_webhook_deliveries(created_at DESC);

-- Composite index for pending deliveries
CREATE INDEX IF NOT EXISTS idx_trust_deliveries_pending ON trust_webhook_deliveries(status, scheduled_at) 
WHERE status = 'pending';

-- Composite index for retries
CREATE INDEX IF NOT EXISTS idx_trust_deliveries_retries ON trust_webhook_deliveries(status, next_retry_at) 
WHERE status = 'retrying';

-- ============================================
-- Trigger for updated_at timestamps
-- ============================================
CREATE OR REPLACE FUNCTION update_trust_webhook_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_trust_webhooks_updated_at ON trust_webhooks;
CREATE TRIGGER update_trust_webhooks_updated_at 
    BEFORE UPDATE ON trust_webhooks 
    FOR EACH ROW 
    EXECUTE FUNCTION update_trust_webhook_updated_at_column();

DROP TRIGGER IF EXISTS update_trust_webhook_deliveries_updated_at ON trust_webhook_deliveries;
CREATE TRIGGER update_trust_webhook_deliveries_updated_at 
    BEFORE UPDATE ON trust_webhook_deliveries 
    FOR EACH ROW 
    EXECUTE FUNCTION update_trust_webhook_updated_at_column();

-- ============================================
-- Comments
-- ============================================
COMMENT ON TABLE trust_webhooks IS 'Webhook configurations for receiving real-time trust event notifications';
COMMENT ON TABLE trust_webhook_deliveries IS 'Delivery attempts and results for trust webhook notifications';
