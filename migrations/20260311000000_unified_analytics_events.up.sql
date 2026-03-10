-- Unified analytics: canonical event store and rollups for Phase 3
-- Migration: 20260311000000_unified_analytics_events.up.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Canonical event store: one row per event (or per aggregated fact from sync job)
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    resource_type TEXT NOT NULL,
    event_type TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    quantity BIGINT NOT NULL DEFAULT 1,
    latency_ms INTEGER,
    cost_usd DECIMAL(12,6),
    resource_id UUID,
    payload JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analytics_events_tenant_id ON analytics_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_occurred_at ON analytics_events(occurred_at);
CREATE INDEX IF NOT EXISTS idx_analytics_events_resource_type ON analytics_events(resource_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_event_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_tenant_occurred ON analytics_events(tenant_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_analytics_events_lookup ON analytics_events(tenant_id, resource_type, event_type, occurred_at);

COMMENT ON TABLE analytics_events IS 'Canonical analytics events; written by call sites or sync job';

-- Pre-aggregated rollups for fast dashboard and admin queries (by tenant, period, metric)
CREATE TABLE IF NOT EXISTS analytics_rollups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    period TEXT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    metric_name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, period, period_start, metric_name)
);

CREATE INDEX IF NOT EXISTS idx_analytics_rollups_tenant_id ON analytics_rollups(tenant_id);
CREATE INDEX IF NOT EXISTS idx_analytics_rollups_period_start ON analytics_rollups(period_start);
CREATE INDEX IF NOT EXISTS idx_analytics_rollups_metric_name ON analytics_rollups(metric_name);
CREATE INDEX IF NOT EXISTS idx_analytics_rollups_lookup ON analytics_rollups(tenant_id, period, metric_name, period_start);

COMMENT ON TABLE analytics_rollups IS 'Pre-computed analytics rollups by tenant, period (hour/day/month), and metric';
COMMENT ON COLUMN analytics_rollups.period IS 'hour, day, or month';
COMMENT ON COLUMN analytics_rollups.metric_name IS 'e.g. function_executions, state_read_ops, agent_calls';
