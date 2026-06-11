-- Function Consciousness: Add retry support for failed deliveries
-- This enables exponential backoff retry for failed notifications

-- Delivery retry queue for failed notifications
CREATE TABLE IF NOT EXISTS consciousness_delivery_retries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insight_id      UUID NOT NULL REFERENCES consciousness_insights(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel         VARCHAR(20) NOT NULL,
    payload         JSONB NOT NULL,
    attempt_count   INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 3,
    next_retry_at   TIMESTAMPTZ NOT NULL,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for finding due retries efficiently
CREATE INDEX IF NOT EXISTS idx_delivery_retries_due 
    ON consciousness_delivery_retries(next_retry_at) 
    WHERE next_retry_at <= NOW();

-- Index for tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_delivery_retries_tenant 
    ON consciousness_delivery_retries(tenant_id, next_retry_at);

-- Index for insight-scoped cleanup
CREATE INDEX IF NOT EXISTS idx_delivery_retries_insight 
    ON consciousness_delivery_retries(insight_id);

COMMENT ON TABLE consciousness_delivery_retries IS 
    'Dead letter queue for failed consciousness notification deliveries with exponential backoff retry';
