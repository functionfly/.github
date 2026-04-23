-- Migration: Create function_webhook_subscriptions table for function deployment event webhooks
-- Description: Adds webhook subscription infrastructure for function deployment events (deployed, failed, scaled, deleted)

CREATE TABLE IF NOT EXISTS function_webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    function_id UUID REFERENCES functions(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret VARCHAR(255) NOT NULL,
    event_types TEXT[] NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_fws_tenant ON function_webhook_subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_fws_function ON function_webhook_subscriptions(function_id);
CREATE INDEX IF NOT EXISTS idx_fws_active ON function_webhook_subscriptions(active);

CREATE TABLE IF NOT EXISTS function_webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES function_webhook_subscriptions(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    response_status INTEGER,
    response_body TEXT,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL DEFAULT false,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_fwd_subscription ON function_webhook_deliveries(subscription_id);
CREATE INDEX IF NOT EXISTS idx_fwd_success ON function_webhook_deliveries(success);
CREATE INDEX IF NOT EXISTS idx_fwd_attempted_at ON function_webhook_deliveries(attempted_at DESC);

COMMENT ON TABLE function_webhook_subscriptions IS 'Webhook subscriptions for function deployment events';
COMMENT ON TABLE function_webhook_deliveries IS 'Delivery attempts and results for function webhook notifications';
