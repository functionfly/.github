-- Drop TimescaleDB and rebuild with plain PostgreSQL + useful extensions
-- Run after: apt install postgresql-17-cron postgresql-17-partman postgresql-17-pg-ivm postgresql-17-rum postgresql-17-pg-uuidv7 postgresql-17-tdigest
-- And: shared_preload_libraries = 'pg_cron,pg_ivm' in postgresql.conf + restart

-- ============================================================
-- Step 1: Drop TimescaleDB hypertables (tables are empty, safe)
-- ============================================================
DROP MATERIALIZED VIEW IF EXISTS signal_stats_hourly CASCADE;
DROP TABLE IF EXISTS signals CASCADE;
DROP TABLE IF EXISTS analytics_events CASCADE;

-- ============================================================
-- Step 2: Create new extensions
-- ============================================================
CREATE EXTENSION IF NOT EXISTS pg_cron;          -- In-DB job scheduling
CREATE EXTENSION IF NOT EXISTS pg_uuidv7;        -- Time-sortable UUIDs
CREATE EXTENSION IF NOT EXISTS tdigest;          -- Fast percentiles
CREATE EXTENSION IF NOT EXISTS rum;              -- Better full-text + timestamp index

-- ============================================================
-- Step 3: Rebuild signals as range-partitioned table
-- ============================================================
CREATE TABLE signals (
    id             UUID NOT NULL DEFAULT uuid_generate_v7(),
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

-- Monthly partitions (create 6 months ahead)
CREATE TABLE signals_2026_06 PARTITION OF signals
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE signals_2026_07 PARTITION OF signals
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE signals_2026_08 PARTITION OF signals
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE signals_2026_09 PARTITION OF signals
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE signals_2026_10 PARTITION OF signals
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE signals_2026_11 PARTITION OF signals
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE signals_2026_12 PARTITION OF signals
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE signals_default PARTITION OF signals DEFAULT;

-- Indexes on partitioned table
CREATE INDEX idx_signals_tenant ON signals(tenant_id, created_at DESC);
CREATE INDEX idx_signals_tenant_type ON signals(tenant_id, signal_type, created_at DESC);
CREATE INDEX idx_signals_tenant_connector ON signals(tenant_id, connector_slug, created_at DESC);
CREATE INDEX idx_signals_entity ON signals(tenant_id, entity_id);
CREATE INDEX idx_signals_created ON signals(created_at DESC);
CREATE INDEX idx_signals_embedding ON signals USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- ============================================================
-- Step 4: Rebuild analytics_events as range-partitioned table
-- ============================================================
CREATE TABLE analytics_events (
    id             UUID NOT NULL DEFAULT uuid_generate_v7(),
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

CREATE TABLE analytics_events_2026_06 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE analytics_events_2026_07 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE analytics_events_2026_08 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE analytics_events_2026_09 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE analytics_events_2026_10 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE analytics_events_2026_11 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE analytics_events_2026_12 PARTITION OF analytics_events
    FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');
CREATE TABLE analytics_events_default PARTITION OF analytics_events DEFAULT;

CREATE INDEX idx_analytics_events_type ON analytics_events(event_type, created_at DESC);
CREATE INDEX idx_analytics_events_tier ON analytics_events(tenant_tier, created_at DESC);
CREATE INDEX idx_analytics_events_connector ON analytics_events(connector_slug, created_at DESC);

-- ============================================================
-- Step 5: Materialized view for signal stats (manual refresh via pg_cron)
-- ============================================================
CREATE MATERIALIZED VIEW signal_stats_hourly AS
SELECT
    date_trunc('hour', created_at) AS bucket,
    tenant_id,
    connector_slug,
    signal_type,
    COUNT(*) as signal_count,
    AVG(importance)::numeric(3,2) as avg_importance
FROM signals
GROUP BY 1, 2, 3, 4;

CREATE INDEX idx_signal_stats_hourly_bucket ON signal_stats_hourly(bucket DESC);

-- ============================================================
-- Step 6: pg_cron jobs (retention + materialized view refresh)
-- ============================================================

-- Refresh signal stats every hour
SELECT cron.schedule('refresh-signal-stats', '0 * * * *', 
    'REFRESH MATERIALIZED VIEW CONCURRENTLY signal_stats_hourly');

-- Drop partitions older than 90 days (runs daily at 3 AM)
SELECT cron.schedule('cleanup-old-signals', '0 3 * * *', $$
    DO $$
    DECLARE
        cutoff DATE := date_trunc('month', CURRENT_DATE - INTERVAL '90 days');
        part_name TEXT;
    BEGIN
        FOR part_name IN
            SELECT inhrelid::regclass::text
            FROM pg_inherits
            JOIN pg_class ON pg_class.oid = inhrelid
            WHERE inhparent = 'signals'::regclass
            AND pg_class.relname ~ '^signals_\d{4}_\d{2}$'
            AND pg_class.relname < 'signals_' || to_char(cutoff, 'YYYY_MM')
        LOOP
            EXECUTE 'DROP TABLE IF EXISTS ' || part_name;
            RAISE NOTICE 'Dropped partition: %', part_name;
        END LOOP;
    END $$;
$$);

-- Drop analytics partitions older than 18 months (runs weekly)
SELECT cron.schedule('cleanup-old-analytics', '0 4 * * 0', $$
    DO $$
    DECLARE
        cutoff DATE := date_trunc('month', CURRENT_DATE - INTERVAL '18 months');
        part_name TEXT;
    BEGIN
        FOR part_name IN
            SELECT inhrelid::regclass::text
            FROM pg_inherits
            JOIN pg_class ON pg_class.oid = inhrelid
            WHERE inhparent = 'analytics_events'::regclass
            AND pg_class.relname ~ '^analytics_events_\d{4}_\d{2}$'
            AND pg_class.relname < 'analytics_events_' || to_char(cutoff, 'YYYY_MM')
        LOOP
            EXECUTE 'DROP TABLE IF EXISTS ' || part_name;
            RAISE NOTICE 'Dropped partition: %', part_name;
        END LOOP;
    END $$;
$$);

-- Auto-create next month's partitions on the 25th
SELECT cron.schedule('create-next-partitions', '0 0 25 * *', $$
    DO $$
    DECLARE
        next_month DATE := date_trunc('month', CURRENT_DATE + INTERVAL '2 months');
        part_name TEXT;
    BEGIN
        -- signals
        part_name := 'signals_' || to_char(next_month, 'YYYY_MM');
        IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = part_name) THEN
            EXECUTE format(
                'CREATE TABLE %I PARTITION OF signals FOR VALUES FROM (%L) TO (%L)',
                part_name, next_month, next_month + INTERVAL '1 month'
            );
            RAISE NOTICE 'Created partition: %', part_name;
        END IF;

        -- analytics_events
        part_name := 'analytics_events_' || to_char(next_month, 'YYYY_MM');
        IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = part_name) THEN
            EXECUTE format(
                'CREATE TABLE %I PARTITION OF analytics_events FOR VALUES FROM (%L) TO (%L)',
                part_name, next_month, next_month + INTERVAL '1 month'
            );
            RAISE NOTICE 'Created partition: %', part_name;
        END IF;
    END $$;
$$);

-- ============================================================
-- Step 7: Drop TimescaleDB extension (no longer needed)
-- ============================================================
-- Uncomment when ready:
-- DROP EXTENSION IF EXISTS timescaledb CASCADE;

-- ============================================================
-- Step 8: Verify
-- ============================================================
SELECT 'Extension' as type, extname as name, extversion as version FROM pg_extension 
WHERE extname IN ('pg_cron','pg_uuidv7','tdigest','rum','vector','pg_trgm','unaccent','pg_stat_statements')
ORDER BY extname;

SELECT 'Partition' as type, tablename as name FROM pg_tables 
WHERE schemaname = 'public' 
AND (tablename LIKE 'signals_%' OR tablename LIKE 'analytics_events_%')
ORDER BY tablename;

SELECT 'Cron Job' as type, jobname as name, schedule FROM cron.job ORDER BY jobname;
