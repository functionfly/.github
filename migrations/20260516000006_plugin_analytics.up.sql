-- Plugin Analytics Table
-- Usage telemetry for plugins

CREATE TABLE IF NOT EXISTS plugin_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plugin_id UUID NOT NULL REFERENCES plugins(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    executions_count INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    total_latency_ms BIGINT DEFAULT 0,
    cpu_usage_seconds DECIMAL(10,2) DEFAULT 0,
    memory_usage_mb_avg DECIMAL(10,2) DEFAULT 0,
    network_bytes BIGINT DEFAULT 0,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_plugin_analytics_plugin ON plugin_analytics(plugin_id);
CREATE INDEX idx_plugin_analytics_tenant ON plugin_analytics(tenant_id);
CREATE INDEX idx_plugin_analytics_period ON plugin_analytics(period_start, period_end);
CREATE INDEX idx_plugin_analytics_event_type ON plugin_analytics(event_type);

COMMENT ON TABLE plugin_analytics IS 'Plugin usage telemetry - executions, errors, latency, resource usage';