-- Migration: Brain signals (partitioned by month with pgvector)
-- Requires: pgvector extension, pg_uuidv7 extension (optional, falls back to gen_random_uuid)

CREATE EXTENSION IF NOT EXISTS vector;

-- uuid_generate_v7() available if pg_uuidv7 installed; otherwise use gen_random_uuid()
DO $$
BEGIN
    CREATE EXTENSION IF NOT EXISTS pg_uuidv7;
EXCEPTION WHEN OTHERS THEN
    NULL;
END $$;

CREATE TABLE IF NOT EXISTS signals (
    id             UUID NOT NULL DEFAULT COALESCE(uuid_generate_v7(), gen_random_uuid()),
    tenant_id      UUID NOT NULL,
    connector_slug VARCHAR(50) NOT NULL,
    signal_type    VARCHAR(100) NOT NULL,
    entity_id      VARCHAR(255),
    entity_name    TEXT,
    fact           TEXT NOT NULL,
    importance     INT NOT NULL DEFAULT 1,
    source_url     TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata       JSONB DEFAULT '{}',
    embedding      VECTOR(1536)
) PARTITION BY RANGE (created_at);

-- Monthly partitions (6 months ahead + default)
CREATE TABLE IF NOT EXISTS signals_2026_06 PARTITION OF signals
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE IF NOT EXISTS signals_2026_07 PARTITION OF signals
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE IF NOT EXISTS signals_2026_08 PARTITION OF signals
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE IF NOT EXISTS signals_2026_09 PARTITION OF signals
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE IF NOT EXISTS signals_2026_10 PARTITION OF signals
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE IF NOT EXISTS signals_2026_11 PARTITION OF signals
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE IF NOT EXISTS signals_2026_12 PARTITION OF signals
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE IF NOT EXISTS signals_default PARTITION OF signals DEFAULT;

CREATE INDEX IF NOT EXISTS idx_signals_tenant ON signals(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_signals_tenant_type ON signals(tenant_id, signal_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_signals_tenant_connector ON signals(tenant_id, connector_slug, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_signals_entity ON signals(tenant_id, entity_id);
CREATE INDEX IF NOT EXISTS idx_signals_created ON signals(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_signals_embedding ON signals USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Brain composers table
CREATE TABLE IF NOT EXISTS brain_composers (
    id             UUID PRIMARY KEY DEFAULT COALESCE(uuid_generate_v7(), gen_random_uuid()),
    tenant_id      UUID NOT NULL,
    name           VARCHAR(255) NOT NULL,
    schedule       VARCHAR(100) NOT NULL DEFAULT '0 8 * * *',
    signal_filters JSONB NOT NULL DEFAULT '[]',
    output_format  VARCHAR(50) NOT NULL DEFAULT 'briefing',
    actions        JSONB NOT NULL DEFAULT '[]',
    is_active      BOOLEAN NOT NULL DEFAULT true,
    last_run_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_brain_composers_tenant ON brain_composers(tenant_id);

-- Brain feedback table
CREATE TABLE IF NOT EXISTS brain_feedback (
    id         UUID PRIMARY KEY DEFAULT COALESCE(uuid_generate_v7(), gen_random_uuid()),
    tenant_id  UUID NOT NULL,
    signal_id  UUID NOT NULL,
    helpful    BOOLEAN NOT NULL,
    context    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_brain_feedback_signal ON brain_feedback(signal_id);
CREATE INDEX IF NOT EXISTS idx_brain_feedback_tenant ON brain_feedback(tenant_id, created_at DESC);
