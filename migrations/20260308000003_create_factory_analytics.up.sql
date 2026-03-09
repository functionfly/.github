-- Create factory analytics tables for tracking performance metrics
-- Migration: 20260308000003_create_factory_analytics.up.sql

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Ensure factory pipeline tables exist (may live in internal/storage/sql/migrations in some setups)
CREATE TABLE IF NOT EXISTS factory_opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL DEFAULT 'automation',
    tags TEXT[] DEFAULT '{}',
    demand_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    complexity INTEGER NOT NULL DEFAULT 1,
    validated BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL DEFAULT 'pending',
    quality_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    review_status TEXT NOT NULL DEFAULT 'not_required',
    review_reason TEXT,
    review_requested_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    generated_func_id TEXT,
    generation_run_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_factory_opportunities_source UNIQUE (source, source_id)
);
CREATE INDEX IF NOT EXISTS idx_factory_opportunities_status ON factory_opportunities(status);
CREATE INDEX IF NOT EXISTS idx_factory_opportunities_source ON factory_opportunities(source);

CREATE TABLE IF NOT EXISTS factory_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    opportunities_scanned INTEGER NOT NULL DEFAULT 0,
    functions_generated INTEGER NOT NULL DEFAULT 0,
    functions_published INTEGER NOT NULL DEFAULT 0,
    average_quality_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    error_message TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_factory_runs_agent_id ON factory_runs(agent_id);
CREATE INDEX IF NOT EXISTS idx_factory_runs_status ON factory_runs(status);

CREATE TABLE IF NOT EXISTS factory_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES factory_runs(id) ON DELETE CASCADE,
    opportunity_id UUID NOT NULL REFERENCES factory_opportunities(id) ON DELETE CASCADE,
    function_id TEXT NOT NULL,
    generated_code TEXT NOT NULL,
    manifest TEXT,
    model_used TEXT,
    quality_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    test_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    review_required BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_factory_versions_run_id ON factory_versions(run_id);
CREATE INDEX IF NOT EXISTS idx_factory_versions_opportunity_id ON factory_versions(opportunity_id);

-- Create enum types for metric types and aggregation periods
DO $$ BEGIN
    CREATE TYPE metric_type AS ENUM (
        'generation_success',
        'generation_failure',
        'quality_score',
        'test_score',
        'latency_generation',
        'latency_testing',
        'latency_publishing',
        'latency_total',
        'error_rate',
        'throughput',
        'opportunity_scanned',
        'function_published',
        'review_required'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE aggregation_period AS ENUM (
        'hourly',
        'daily',
        'weekly',
        'monthly'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create factory_analytics_metrics table for raw metric data points
CREATE TABLE IF NOT EXISTS factory_analytics_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID REFERENCES factory_runs(id) ON DELETE SET NULL,
    agent_id TEXT NOT NULL,
    metric_type TEXT NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    labels JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_factory_analytics_metrics_run_id ON factory_analytics_metrics(run_id);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_metrics_agent_id ON factory_analytics_metrics(agent_id);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_metrics_metric_type ON factory_analytics_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_metrics_created_at ON factory_analytics_metrics(created_at);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_metrics_agent_type ON factory_analytics_metrics(agent_id, metric_type);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_metrics_agent_created ON factory_analytics_metrics(agent_id, created_at);

-- Create composite index for time-series queries
CREATE INDEX IF NOT EXISTS idx_factory_analytics_metrics_timeseries ON factory_analytics_metrics(agent_id, metric_type, created_at DESC);

-- Create GIN index for JSONB labels queries
CREATE INDEX IF NOT EXISTS idx_factory_analytics_metrics_labels ON factory_analytics_metrics USING GIN (labels);

-- Create factory_analytics_aggregated table for pre-computed statistics
CREATE TABLE IF NOT EXISTS factory_analytics_aggregated (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL,
    period TEXT NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    metric_type TEXT NOT NULL,
    count BIGINT NOT NULL DEFAULT 0,
    sum DOUBLE PRECISION DEFAULT 0,
    avg DOUBLE PRECISION DEFAULT 0,
    min DOUBLE PRECISION DEFAULT 0,
    max DOUBLE PRECISION DEFAULT 0,
    p50 DOUBLE PRECISION DEFAULT 0,
    p95 DOUBLE PRECISION DEFAULT 0,
    p99 DOUBLE PRECISION DEFAULT 0,
    success_count BIGINT DEFAULT 0,
    failure_count BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(agent_id, period, period_start, metric_type)
);

-- Create indexes for aggregated metrics
CREATE INDEX IF NOT EXISTS idx_factory_analytics_aggregated_agent_id ON factory_analytics_aggregated(agent_id);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_aggregated_period ON factory_analytics_aggregated(period);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_aggregated_period_start ON factory_analytics_aggregated(period_start);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_aggregated_metric_type ON factory_analytics_aggregated(metric_type);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_aggregated_agent_period ON factory_analytics_aggregated(agent_id, period);
CREATE INDEX IF NOT EXISTS idx_factory_analytics_aggregated_lookup ON factory_analytics_aggregated(agent_id, period, metric_type, period_start);

-- Create a view for dashboard statistics
CREATE OR REPLACE VIEW factory_analytics_dashboard AS
SELECT
    agent_id,
    NOW() AS last_updated,
    -- Run statistics (last 24 hours)
    (SELECT COUNT(*) FROM factory_runs fr WHERE fr.agent_id = fa.agent_id AND fr.created_at >= NOW() - INTERVAL '24 hours') AS runs_24h,
    (SELECT COUNT(*) FROM factory_runs fr WHERE fr.agent_id = fa.agent_id AND fr.status = 'succeeded' AND fr.created_at >= NOW() - INTERVAL '24 hours') AS successful_runs_24h,
    (SELECT COUNT(*) FROM factory_runs fr WHERE fr.agent_id = fa.agent_id AND fr.status = 'failed' AND fr.created_at >= NOW() - INTERVAL '24 hours') AS failed_runs_24h,
    -- Quality metrics (last 24 hours)
    (SELECT COALESCE(AVG(fv.quality_score), 0) FROM factory_versions fv JOIN factory_runs fr ON fv.run_id = fr.id WHERE fr.agent_id = fa.agent_id AND fv.created_at >= NOW() - INTERVAL '24 hours') AS avg_quality_24h,
    (SELECT COALESCE(AVG(fv.test_score), 0) FROM factory_versions fv JOIN factory_runs fr ON fv.run_id = fr.id WHERE fr.agent_id = fa.agent_id AND fv.created_at >= NOW() - INTERVAL '24 hours') AS avg_test_24h,
    -- Throughput (last 24 hours)
    (SELECT COUNT(*) FROM factory_versions fv JOIN factory_runs fr ON fv.run_id = fr.id WHERE fr.agent_id = fa.agent_id AND fv.created_at >= NOW() - INTERVAL '24 hours') AS functions_generated_24h,
    (SELECT COUNT(*) FROM factory_versions fv JOIN factory_runs fr ON fv.run_id = fr.id WHERE fr.agent_id = fa.agent_id AND NOT fv.review_required AND fv.created_at >= NOW() - INTERVAL '24 hours') AS functions_published_24h,
    -- Review metrics (last 24 hours)
    (SELECT COUNT(*) FROM factory_versions fv JOIN factory_runs fr ON fv.run_id = fr.id WHERE fr.agent_id = fa.agent_id AND fv.review_required AND fv.created_at >= NOW() - INTERVAL '24 hours') AS pending_reviews_24h
FROM (SELECT DISTINCT agent_id FROM factory_runs) fa;

-- Create a function to aggregate metrics for a time period
CREATE OR REPLACE FUNCTION aggregate_factory_metrics(
    p_agent_id TEXT,
    p_period TEXT,
    p_period_start TIMESTAMPTZ
) RETURNS VOID AS $$
DECLARE
    v_period_end TIMESTAMPTZ;
    v_metric_type TEXT;
    v_count BIGINT;
    v_sum DOUBLE PRECISION;
    v_avg DOUBLE PRECISION;
    v_min DOUBLE PRECISION;
    v_max DOUBLE PRECISION;
    v_p50 DOUBLE PRECISION;
    v_p95 DOUBLE PRECISION;
    v_p99 DOUBLE PRECISION;
    v_success_count BIGINT;
    v_failure_count BIGINT;
BEGIN
    -- Calculate period end based on period type
    CASE p_period
        WHEN 'hourly' THEN v_period_end := p_period_start + INTERVAL '1 hour';
        WHEN 'daily' THEN v_period_end := p_period_start + INTERVAL '1 day';
        WHEN 'weekly' THEN v_period_end := p_period_start + INTERVAL '7 days';
        WHEN 'monthly' THEN v_period_end := p_period_start + INTERVAL '1 month';
        ELSE v_period_end := p_period_start + INTERVAL '1 hour';
    END CASE;

    -- Aggregate each metric type
    FOR v_metric_type IN
        SELECT DISTINCT metric_type
        FROM factory_analytics_metrics
        WHERE agent_id = p_agent_id
        AND created_at >= p_period_start
        AND created_at < v_period_end
    LOOP
        -- Calculate statistics
        SELECT
            COUNT(*),
            COALESCE(SUM(metric_value), 0),
            COALESCE(AVG(metric_value), 0),
            COALESCE(MIN(metric_value), 0),
            COALESCE(MAX(metric_value), 0),
            COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY metric_value), 0),
            COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY metric_value), 0),
            COALESCE(percentile_cont(0.99) WITHIN GROUP (ORDER BY metric_value), 0)
        INTO v_count, v_sum, v_avg, v_min, v_max, v_p50, v_p95, v_p99
        FROM factory_analytics_metrics
        WHERE agent_id = p_agent_id
        AND metric_type = v_metric_type
        AND created_at >= p_period_start
        AND created_at < v_period_end;

        -- Get success/failure counts for generation metrics
        IF v_metric_type IN ('generation_success', 'generation_failure') THEN
            SELECT
                COUNT(*) FILTER (WHERE metric_type = 'generation_success'),
                COUNT(*) FILTER (WHERE metric_type = 'generation_failure')
            INTO v_success_count, v_failure_count
            FROM factory_analytics_metrics
            WHERE agent_id = p_agent_id
            AND metric_type IN ('generation_success', 'generation_failure')
            AND created_at >= p_period_start
            AND created_at < v_period_end;
        ELSE
            v_success_count := 0;
            v_failure_count := 0;
        END IF;

        -- Upsert aggregated record
        INSERT INTO factory_analytics_aggregated (
            agent_id, period, period_start, metric_type,
            count, sum, avg, min, max, p50, p95, p99,
            success_count, failure_count, created_at, updated_at
        ) VALUES (
            p_agent_id, p_period, p_period_start, v_metric_type,
            v_count, v_sum, v_avg, v_min, v_max, v_p50, v_p95, v_p99,
            v_success_count, v_failure_count, NOW(), NOW()
        )
        ON CONFLICT (agent_id, period, period_start, metric_type)
        DO UPDATE SET
            count = EXCLUDED.count,
            sum = EXCLUDED.sum,
            avg = EXCLUDED.avg,
            min = EXCLUDED.min,
            max = EXCLUDED.max,
            p50 = EXCLUDED.p50,
            p95 = EXCLUDED.p95,
            p99 = EXCLUDED.p99,
            success_count = EXCLUDED.success_count,
            failure_count = EXCLUDED.failure_count,
            updated_at = NOW();
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Create a function to clean up old metrics
CREATE OR REPLACE FUNCTION cleanup_old_factory_metrics(
    p_retention_days INTEGER DEFAULT 90
) RETURNS BIGINT AS $$
DECLARE
    v_deleted BIGINT;
BEGIN
    DELETE FROM factory_analytics_metrics
    WHERE created_at < NOW() - (p_retention_days || ' days')::INTERVAL;

    GET DIAGNOSTICS v_deleted = ROW_COUNT;
    RETURN v_deleted;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to automatically aggregate metrics after insert
-- Note: This is optional and can be disabled for high-throughput scenarios
-- where aggregation is done via a scheduled job instead

-- Create a view for recent run metrics
CREATE OR REPLACE VIEW factory_analytics_recent_runs AS
SELECT
    fr.id AS run_id,
    fr.agent_id,
    fr.status,
    EXTRACT(EPOCH FROM (COALESCE(fr.completed_at, NOW()) - fr.created_at)) * 1000 AS duration_ms,
    fr.opportunities_scanned,
    fr.functions_generated,
    fr.functions_published,
    fr.average_quality_score,
    fr.created_at,
    fr.completed_at,
    -- Calculate metrics from factory_versions
    (SELECT COALESCE(AVG(fv.test_score), 0) FROM factory_versions fv WHERE fv.run_id = fr.id) AS avg_test_score,
    (SELECT COUNT(*) FROM factory_versions fv WHERE fv.run_id = fr.id AND fv.review_required) AS review_required_count
FROM factory_runs fr
ORDER BY fr.created_at DESC
LIMIT 100;

-- Add comments to tables
COMMENT ON TABLE factory_analytics_metrics IS 'Raw metric data points for factory performance tracking';
COMMENT ON TABLE factory_analytics_aggregated IS 'Pre-computed aggregated statistics for dashboard queries';
COMMENT ON COLUMN factory_analytics_metrics.metric_type IS 'Type of metric: generation_success, quality_score, latency_total, etc.';
COMMENT ON COLUMN factory_analytics_metrics.metric_value IS 'The numeric value of the metric';
COMMENT ON COLUMN factory_analytics_metrics.labels IS 'Additional metadata as JSONB key-value pairs';
COMMENT ON COLUMN factory_analytics_aggregated.period IS 'Aggregation period: hourly, daily, weekly, monthly';
COMMENT ON COLUMN factory_analytics_aggregated.p50 IS '50th percentile (median) value';
COMMENT ON COLUMN factory_analytics_aggregated.p95 IS '95th percentile value';
COMMENT ON COLUMN factory_analytics_aggregated.p99 IS '99th percentile value';

-- Grant permissions (adjust as needed for your security model)
-- GRANT SELECT, INSERT ON factory_analytics_metrics TO factory_app;
-- GRANT SELECT, INSERT, UPDATE ON factory_analytics_aggregated TO factory_app;
-- GRANT EXECUTE ON FUNCTION aggregate_factory_metrics TO factory_app;
-- GRANT EXECUTE ON FUNCTION cleanup_old_factory_metrics TO factory_app;
