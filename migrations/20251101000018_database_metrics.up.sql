-- Database metrics table for storing historical database performance data
-- This enables time-series analysis and growth rate calculations

CREATE TABLE IF NOT EXISTS database_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_type VARCHAR(50) NOT NULL, -- 'connections', 'size_gb', 'query_time', 'cache_hit_ratio', 'throughput'
    value DECIMAL(15,6) NOT NULL,
    unit VARCHAR(20) NOT NULL, -- 'count', 'gb', 'ms', 'ratio', 'qps'
    metadata JSONB DEFAULT '{}', -- Additional context like active/idle breakdown, etc.
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_database_metrics_type_recorded ON database_metrics(metric_type, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_database_metrics_recorded ON database_metrics(recorded_at DESC);

-- Retention policy: keep data for 90 days
CREATE OR REPLACE FUNCTION cleanup_old_database_metrics()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM database_metrics
    WHERE recorded_at < NOW() - INTERVAL '90 days';

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;