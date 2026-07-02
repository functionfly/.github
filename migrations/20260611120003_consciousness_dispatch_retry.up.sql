-- Function Consciousness: Add dispatch retry table for failed notification delivery
-- This enables retry with exponential backoff for failed email/Slack/webhook dispatches

CREATE TABLE IF NOT EXISTS consciousness_dispatch_retry (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insight_id      UUID NOT NULL REFERENCES consciousness_insights(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel         VARCHAR(20) NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    attempt_count   INT NOT NULL DEFAULT 1,
    next_retry_at   TIMESTAMPTZ NOT NULL,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_insight_channel UNIQUE (insight_id, channel)
);

CREATE INDEX IF NOT EXISTS idx_dispatch_retry_next_retry ON consciousness_dispatch_retry(next_retry_at) WHERE attempt_count < 3;
CREATE INDEX IF NOT EXISTS idx_dispatch_retry_tenant ON consciousness_dispatch_retry(tenant_id, next_retry_at);