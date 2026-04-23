-- Migration: TimescaleDB Hypertable Setup for Time-Series Data
-- Creates hypertables for high-volume time-series tables with automatic partitioning
-- and advanced time-series features (compression, retention policies, continuous aggregates)
-- Created: 2026-04-19

-- ============================================
-- Pre-requisite Check: Ensure TimescaleDB Extension is Available
-- ============================================
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        RAISE EXCEPTION 'TimescaleDB extension not found. Please install TimescaleDB first.';
    END IF;
END $$;

-- ============================================
-- 1. Cost Allocation Entries Hypertable
-- High-volume billing data with time-series patterns
-- ============================================

-- First, create the hypertable-compatible table structure
-- Note: TimescaleDB requires unique constraints to include the time column
CREATE TABLE IF NOT EXISTS cost_allocation_entries_ts (
    id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    api_key_id UUID,
    function_id UUID NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    function_author VARCHAR(255) NOT NULL,
    execution_id UUID NOT NULL,
    execution_outcome VARCHAR(50) NOT NULL,
    cached BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Resource usage
    duration_ms BIGINT NOT NULL,
    cpu_time_ms BIGINT,
    memory_used_mb BIGINT,
    wall_time_ms BIGINT,
    
    -- Cost breakdown (in cents)
    execution_cost_cents BIGINT NOT NULL,
    compute_cost_cents BIGINT,
    platform_fee_cents BIGINT,
    data_transfer_cents BIGINT,
    total_cost_cents BIGINT NOT NULL,
    
    -- Metadata
    region VARCHAR(50),
    timestamp TIMESTAMPTZ NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    tags JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Composite primary key including timestamp (TimescaleDB requirement)
    PRIMARY KEY (id, timestamp)
);

-- Convert to hypertable with daily chunks (optimize for ~1M rows/day)
-- For retention alignment, use 1 day chunks
SELECT create_hypertable('cost_allocation_entries_ts', 'timestamp', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Enable compression for chunks older than 7 days
-- Compresses cost data by ~70-80%
ALTER TABLE cost_allocation_entries_ts SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'tenant_id, function_id',
    timescaledb.compress_orderby = 'timestamp DESC'
);

-- Add compression policy: compress after 7 days
SELECT add_compression_policy('cost_allocation_entries_ts', INTERVAL '7 days');

-- Add retention policy: drop after 90 days (configurable via settings)
SELECT add_retention_policy('cost_allocation_entries_ts', INTERVAL '90 days');

-- ============================================
-- 2. Registry Function Executions Hypertable
-- High-volume execution tracking
-- ============================================

CREATE TABLE IF NOT EXISTS registry_function_executions_ts (
    id UUID NOT NULL,
    function_id UUID NOT NULL,
    version VARCHAR(255) NOT NULL,
    duration_ms INTEGER NOT NULL,
    status_code INTEGER NOT NULL,
    cached BOOLEAN NOT NULL DEFAULT FALSE,
    outcome VARCHAR(50) NOT NULL,
    error_code TEXT,
    caller_ip VARCHAR(45),
    user_agent TEXT,
    geo_country VARCHAR(2),
    tenant_id UUID,
    user_id UUID,
    timestamp TIMESTAMPTZ NOT NULL,
    verification_status TEXT,
    verification_error TEXT,
    replayed_duration_ms INTEGER,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, timestamp)
);

SELECT create_hypertable('registry_function_executions_ts', 'timestamp',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

ALTER TABLE registry_function_executions_ts SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'function_id, tenant_id',
    timescaledb.compress_orderby = 'timestamp DESC'
);

SELECT add_compression_policy('registry_function_executions_ts', INTERVAL '7 days');
SELECT add_retention_policy('registry_function_executions_ts', INTERVAL '30 days');

-- ============================================
-- 3. Health Checks Hypertable
-- System monitoring data
-- ============================================

CREATE TABLE IF NOT EXISTS health_checks_ts (
    id UUID NOT NULL,
    backend_id UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    ok BOOLEAN NOT NULL,
    status_code INTEGER,
    latency_ms INTEGER,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, timestamp)
);

SELECT create_hypertable('health_checks_ts', 'timestamp',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

ALTER TABLE health_checks_ts SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'backend_id',
    timescaledb.compress_orderby = 'timestamp DESC'
);

SELECT add_compression_policy('health_checks_ts', INTERVAL '7 days');
SELECT add_retention_policy('health_checks_ts', INTERVAL '90 days');

-- ============================================
-- 4. Routing Events Hypertable
-- Traffic routing and observability data
-- ============================================

CREATE TABLE IF NOT EXISTS routing_events_ts (
    id UUID NOT NULL,
    app_id UUID NOT NULL,
    backend_id UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    latency_ms INTEGER,
    outcome VARCHAR(20) NOT NULL,
    request_id VARCHAR(255),
    region VARCHAR(50),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, timestamp)
);

SELECT create_hypertable('routing_events_ts', 'timestamp',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

ALTER TABLE routing_events_ts SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'app_id, backend_id',
    timescaledb.compress_orderby => 'timestamp DESC'
);

SELECT add_compression_policy('routing_events_ts', INTERVAL '7 days');
SELECT add_retention_policy('routing_events_ts', INTERVAL '30 days');

-- ============================================
-- 5. Performance Metrics Hypertable
-- Detailed performance tracking
-- ============================================

CREATE TABLE IF NOT EXISTS performance_metrics_ts (
    id UUID NOT NULL,
    metric_type VARCHAR(100) NOT NULL,
    tenant_id UUID,
    app_id UUID,
    backend_id UUID,
    value DECIMAL(20,6) NOT NULL,
    unit VARCHAR(20),
    labels JSONB,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, timestamp)
);

SELECT create_hypertable('performance_metrics_ts', 'timestamp',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

ALTER TABLE performance_metrics_ts SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'metric_type, tenant_id',
    timescaledb.compress_orderby = 'timestamp DESC'
);

SELECT add_compression_policy('performance_metrics_ts', INTERVAL '7 days');
SELECT add_retention_policy('performance_metrics_ts', INTERVAL '30 days');

-- ============================================
-- 6. Continuous Aggregates for Real-Time Analytics
-- Pre-computed rollups for dashboard queries
-- ============================================

-- Hourly cost aggregation by tenant
CREATE MATERIALIZED VIEW cost_hourly_summary
WITH (timescaledb.continuous) AS
SELECT 
    tenant_id,
    time_bucket('1 hour', timestamp) AS bucket,
    function_id,
    COUNT(*) as execution_count,
    SUM(total_cost_cents) as total_cost_cents,
    AVG(duration_ms) as avg_duration_ms,
    SUM(CASE WHEN execution_outcome = 'success' THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN cached THEN 1 ELSE 0 END) as cached_count
FROM cost_allocation_entries_ts
GROUP BY tenant_id, bucket, function_id;

-- Policy to auto-refresh hourly (last 7 days)
SELECT add_continuous_aggregate_policy('cost_hourly_summary',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour'
);

-- Daily tenant summary (longer retention)
CREATE MATERIALIZED VIEW cost_daily_summary
WITH (timescaledb.continuous) AS
SELECT 
    tenant_id,
    time_bucket('1 day', timestamp) AS bucket,
    COUNT(*) as execution_count,
    SUM(total_cost_cents) as total_cost_cents,
    SUM(execution_cost_cents) as execution_cost_cents,
    SUM(compute_cost_cents) as compute_cost_cents,
    SUM(platform_fee_cents) as platform_fee_cents,
    SUM(data_transfer_cents) as data_transfer_cents,
    COUNT(DISTINCT function_id) as unique_functions,
    AVG(duration_ms) as avg_duration_ms,
    MAX(timestamp) as last_execution
FROM cost_allocation_entries_ts
GROUP BY tenant_id, bucket;

SELECT add_continuous_aggregate_policy('cost_daily_summary',
    start_offset => INTERVAL '1 year',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day'
);

-- Hourly execution summary for dashboard
CREATE MATERIALIZED VIEW execution_hourly_summary
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', timestamp) AS bucket,
    tenant_id,
    function_id,
    COUNT(*) as execution_count,
    SUM(CASE WHEN outcome = 'success' THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN outcome = 'error' THEN 1 ELSE 0 END) as error_count,
    AVG(duration_ms) as avg_duration_ms,
    MAX(duration_ms) as max_duration_ms,
    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms) as p95_duration_ms,
    PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms) as p99_duration_ms
FROM registry_function_executions_ts
GROUP BY bucket, tenant_id, function_id;

SELECT add_continuous_aggregate_policy('execution_hourly_summary',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour'
);

-- ============================================
-- 7. Indexes for Hypertables
-- ============================================

-- Cost allocation indexes
CREATE INDEX idx_cost_allocation_ts_tenant_time 
ON cost_allocation_entries_ts (tenant_id, timestamp DESC);

CREATE INDEX idx_cost_allocation_ts_function_time 
ON cost_allocation_entries_ts (function_id, timestamp DESC);

CREATE INDEX idx_cost_allocation_ts_outcome 
ON cost_allocation_entries_ts (execution_outcome, timestamp DESC) 
WHERE execution_outcome != 'success';

-- Execution indexes
CREATE INDEX idx_executions_ts_function_time 
ON registry_function_executions_ts (function_id, timestamp DESC);

CREATE INDEX idx_executions_ts_tenant_time 
ON registry_function_executions_ts (tenant_id, timestamp DESC) 
WHERE tenant_id IS NOT NULL;

CREATE INDEX idx_executions_ts_outcome_time 
ON registry_function_executions_ts (outcome, timestamp DESC);

-- Health check indexes
CREATE INDEX idx_health_checks_ts_backend_time 
ON health_checks_ts (backend_id, timestamp DESC);

CREATE INDEX idx_health_checks_ts_ok_time 
ON health_checks_ts (ok, timestamp DESC) 
WHERE NOT ok;

-- Routing indexes
CREATE INDEX idx_routing_ts_app_time 
ON routing_events_ts (app_id, timestamp DESC);

CREATE INDEX idx_routing_ts_backend_time 
ON routing_events_ts (backend_id, timestamp DESC);

CREATE INDEX idx_routing_ts_outcome 
ON routing_events_ts (outcome, timestamp DESC);

-- ============================================
-- 8. Migration Functions: Copy from Old Tables
-- ============================================

-- Function to migrate data from old tables to hypertables
CREATE OR REPLACE FUNCTION migrate_to_timescale_tables(
    p_batch_size INTEGER DEFAULT 10000,
    p_max_age_days INTEGER DEFAULT 30
)
RETURNS TABLE (table_name TEXT, rows_migrated BIGINT) AS $$
DECLARE
    v_cutoff TIMESTAMPTZ;
    v_count BIGINT;
BEGIN
    v_cutoff := NOW() - (p_max_age_days || ' days')::INTERVAL;
    
    -- Migrate cost_allocation_entries (recent data only for performance)
    INSERT INTO cost_allocation_entries_ts (
        id, tenant_id, api_key_id, function_id, function_name, function_author,
        execution_id, execution_outcome, cached, duration_ms, cpu_time_ms,
        memory_used_mb, wall_time_ms, execution_cost_cents, compute_cost_cents,
        platform_fee_cents, data_transfer_cents, total_cost_cents, region,
        timestamp, period_start, period_end, tags, metadata, created_at
    )
    SELECT 
        id, tenant_id, api_key_id, function_id, function_name, function_author,
        execution_id, execution_outcome, cached, duration_ms, cpu_time_ms,
        memory_used_mb, wall_time_ms, execution_cost_cents, compute_cost_cents,
        platform_fee_cents, data_transfer_cents, total_cost_cents, region,
        timestamp, period_start, period_end, tags, metadata, created_at
    FROM cost_allocation_entries
    WHERE timestamp >= v_cutoff
    ON CONFLICT (id, timestamp) DO NOTHING;
    
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN QUERY SELECT 'cost_allocation_entries_ts'::TEXT, v_count;
    
    -- Migrate registry_function_executions
    INSERT INTO registry_function_executions_ts (
        id, function_id, version, duration_ms, status_code, cached, outcome,
        error_code, caller_ip, user_agent, geo_country, tenant_id, user_id,
        timestamp, verification_status, verification_error, replayed_duration_ms,
        metadata, created_at
    )
    SELECT 
        id, function_id, version, duration_ms, status_code, cached, outcome,
        error_code, caller_ip, user_agent, geo_country, tenant_id, user_id,
        timestamp, verification_status, verification_error, replayed_duration_ms,
        '{}'::jsonb, created_at
    FROM registry_function_executions
    WHERE timestamp >= v_cutoff
    ON CONFLICT (id, timestamp) DO NOTHING;
    
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN QUERY SELECT 'registry_function_executions_ts'::TEXT, v_count;
    
    -- Migrate health_checks
    INSERT INTO health_checks_ts (
        id, backend_id, timestamp, ok, status_code, latency_ms, error_message,
        metadata, created_at
    )
    SELECT 
        id, backend_id, timestamp, ok, status_code, latency_ms, error_message,
        '{}'::jsonb, created_at
    FROM health_checks
    WHERE timestamp >= v_cutoff
    ON CONFLICT (id, timestamp) DO NOTHING;
    
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN QUERY SELECT 'health_checks_ts'::TEXT, v_count;
    
    -- Migrate routing_events
    INSERT INTO routing_events_ts (
        id, app_id, backend_id, timestamp, latency_ms, outcome, request_id,
        region, metadata, created_at
    )
    SELECT 
        id, app_id, backend_id, timestamp, latency_ms, outcome, request_id,
        NULL, '{}'::jsonb, created_at
    FROM routing_events
    WHERE timestamp >= v_cutoff
    ON CONFLICT (id, timestamp) DO NOTHING;
    
    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN QUERY SELECT 'routing_events_ts'::TEXT, v_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 9. Real-Time Data Policies
-- ============================================

-- Function to customize retention per tenant (for enterprise customers)
CREATE OR REPLACE FUNCTION set_tenant_retention_policy(
    p_tenant_id UUID,
    p_retention_days INTEGER
)
RETURNS VOID AS $$
BEGIN
    -- Note: TimescaleDB retention policies are global per table
    -- For per-tenant retention, we need custom cleanup logic
    -- This creates a scheduled job reference for the cleanup task
    
    INSERT INTO tenant_retention_policies (tenant_id, retention_days, updated_at)
    VALUES (p_tenant_id, p_retention_days, NOW())
    ON CONFLICT (tenant_id) DO UPDATE SET
        retention_days = p_retention_days,
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Create the tenant retention policies tracking table
CREATE TABLE IF NOT EXISTS tenant_retention_policies (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    retention_days INTEGER NOT NULL DEFAULT 90,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE tenant_retention_policies IS 'Custom retention policies per tenant. Used by custom cleanup jobs.';

-- ============================================
-- 10. Comments for Documentation
-- ============================================

COMMENT ON TABLE cost_allocation_entries_ts IS 'TimescaleDB hypertable for cost allocation data. Daily chunks, 7-day compression, 90-day retention.';
COMMENT ON TABLE registry_function_executions_ts IS 'TimescaleDB hypertable for function executions. Daily chunks, 7-day compression, 30-day retention.';
COMMENT ON TABLE health_checks_ts IS 'TimescaleDB hypertable for health checks. Daily chunks, 7-day compression, 90-day retention.';
COMMENT ON TABLE routing_events_ts IS 'TimescaleDB hypertable for routing events. Daily chunks, 7-day compression, 30-day retention.';
COMMENT ON TABLE performance_metrics_ts IS 'TimescaleDB hypertable for performance metrics. Daily chunks, 7-day compression, 30-day retention.';

COMMENT ON MATERIALIZED VIEW cost_hourly_summary IS 'TimescaleDB continuous aggregate. Hourly cost data by tenant/function. Auto-refreshes hourly.';
COMMENT ON MATERIALIZED VIEW cost_daily_summary IS 'TimescaleDB continuous aggregate. Daily cost data by tenant. Auto-refreshes daily.';
COMMENT ON MATERIALIZED VIEW execution_hourly_summary IS 'TimescaleDB continuous aggregate. Hourly execution stats with percentiles. Auto-refreshes hourly.';
