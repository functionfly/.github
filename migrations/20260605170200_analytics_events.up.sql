-- Migration: Analytics events (partitioned by month)
-- Anonymized, privacy-safe events for model improvement

CREATE TABLE IF NOT EXISTS analytics_events (
    id             UUID NOT NULL DEFAULT COALESCE(uuid_generate_v7(), gen_random_uuid()),
    event_type     VARCHAR(50) NOT NULL,
    tenant_tier    VARCHAR(20) NOT NULL,
    connector_slug VARCHAR(50),
    signal_type    VARCHAR(100),
    importance     INT,
    signals_count  INT,
    fact_length    INT,
    metadata       JSONB DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);

CREATE TABLE IF NOT EXISTS analytics_events_2026_06 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE IF NOT EXISTS analytics_events_2026_07 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE IF NOT EXISTS analytics_events_2026_08 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE IF NOT EXISTS analytics_events_2026_09 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE IF NOT EXISTS analytics_events_2026_10 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE IF NOT EXISTS analytics_events_2026_11 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE IF NOT EXISTS analytics_events_2026_12 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE IF NOT EXISTS analytics_events_default PARTITION OF analytics_events DEFAULT;

CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_events_tier ON analytics_events(tenant_tier, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_events_connector ON analytics_events(connector_slug, created_at DESC);

-- Brain trigger configs (daemon triggers on signal patterns)
CREATE TABLE IF NOT EXISTS brain_triggers (
    id              UUID PRIMARY KEY DEFAULT COALESCE(uuid_generate_v7(), gen_random_uuid()),
    tenant_id       UUID NOT NULL,
    agent_id        UUID,
    name            VARCHAR(255) NOT NULL,
    signal_types    TEXT[] NOT NULL DEFAULT '{}',
    connector_slugs TEXT[] NOT NULL DEFAULT '{}',
    min_importance  INT NOT NULL DEFAULT 1,
    schedule        VARCHAR(100) NOT NULL DEFAULT 'immediate',
    action          VARCHAR(50) NOT NULL DEFAULT 'run_agent',
    action_config   JSONB DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    last_fired_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_brain_triggers_tenant ON brain_triggers(tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_brain_triggers_agent ON brain_triggers(agent_id) WHERE agent_id IS NOT NULL;
