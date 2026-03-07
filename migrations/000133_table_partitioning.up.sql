-- Time-based table partitioning for high-volume tables
-- This migration sets up monthly partitioning for tables with high insert/update rates

-- ============================================
-- Registry Function Executions Partitioning
-- ============================================

-- Create partitioned table for registry function executions
-- This table receives high volume inserts and is queried by time ranges frequently

-- First, create the partitioned table structure
CREATE TABLE IF NOT EXISTS registry_function_executions_partitioned (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
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
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    verification_status TEXT,
    verification_error TEXT,
    replayed_duration_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

-- Create indexes on the partitioned table (these will be inherited by partitions)
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_partitioned_function_timestamp
ON registry_function_executions_partitioned(function_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_registry_function_executions_partitioned_tenant_timestamp
ON registry_function_executions_partitioned(tenant_id, timestamp DESC)
WHERE tenant_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_registry_function_executions_partitioned_outcome_timestamp
ON registry_function_executions_partitioned(outcome, timestamp DESC);

-- Create monthly partitions for the current and next few months
-- Current month
CREATE TABLE IF NOT EXISTS registry_function_executions_y2025m01 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2025m02 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2025m03 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

-- Create default partition for future data
CREATE TABLE IF NOT EXISTS registry_function_executions_default PARTITION OF registry_function_executions_partitioned
DEFAULT;

-- ============================================
-- Performance Metrics Partitioning
-- ============================================

-- Create partitioned table for performance metrics
CREATE TABLE IF NOT EXISTS performance_metrics_partitioned (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    metric_type VARCHAR(100) NOT NULL,
    tenant_id UUID,
    app_id UUID,
    backend_id UUID,
    value DECIMAL(20,6) NOT NULL,
    unit VARCHAR(20),
    labels JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

-- Create indexes on the partitioned table
CREATE INDEX IF NOT EXISTS idx_performance_metrics_partitioned_type_timestamp
ON performance_metrics_partitioned(metric_type, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_performance_metrics_partitioned_tenant_timestamp
ON performance_metrics_partitioned(tenant_id, timestamp DESC)
WHERE tenant_id IS NOT NULL;

-- Create monthly partitions
CREATE TABLE IF NOT EXISTS performance_metrics_y2025m01 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2025m02 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2025m03 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE TABLE IF NOT EXISTS performance_metrics_default PARTITION OF performance_metrics_partitioned
DEFAULT;

-- ============================================
-- System Health Checks Partitioning
-- ============================================

-- Create partitioned table for system health checks
CREATE TABLE IF NOT EXISTS system_health_checks_partitioned (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    check_type VARCHAR(100) NOT NULL,
    component_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    response_time_ms INTEGER,
    message TEXT,
    metadata JSONB,
    checked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (checked_at);

-- Create indexes on the partitioned table
CREATE INDEX IF NOT EXISTS idx_system_health_checks_partitioned_type_checked
ON system_health_checks_partitioned(check_type, checked_at DESC);

CREATE INDEX IF NOT EXISTS idx_system_health_checks_partitioned_status_checked
ON system_health_checks_partitioned(status, checked_at DESC);

-- Create monthly partitions
CREATE TABLE IF NOT EXISTS system_health_checks_y2025m01 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2025m02 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2025m03 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE TABLE IF NOT EXISTS system_health_checks_default PARTITION OF system_health_checks_partitioned
DEFAULT;

-- ============================================
-- Database Metrics Partitioning
-- ============================================

-- Create partitioned table for database metrics
CREATE TABLE IF NOT EXISTS database_metrics_partitioned (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    metric_type VARCHAR(100) NOT NULL,
    value DECIMAL(20,6) NOT NULL,
    unit VARCHAR(20),
    labels JSONB,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (recorded_at);

-- Create indexes on the partitioned table
CREATE INDEX IF NOT EXISTS idx_database_metrics_partitioned_type_recorded
ON database_metrics_partitioned(metric_type, recorded_at DESC);

-- Create monthly partitions
CREATE TABLE IF NOT EXISTS database_metrics_y2025m01 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2025m02 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2025m03 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE TABLE IF NOT EXISTS database_metrics_default PARTITION OF database_metrics_partitioned
DEFAULT;

-- ============================================
-- Function Logs Partitioning (High Volume)
-- ============================================

-- Create partitioned table for function logs
CREATE TABLE IF NOT EXISTS function_logs_partitioned (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    function_id UUID,
    deployment_id UUID,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (timestamp);

-- Create indexes on the partitioned table
CREATE INDEX IF NOT EXISTS idx_function_logs_partitioned_function_timestamp
ON function_logs_partitioned(function_id, timestamp DESC)
WHERE function_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_function_logs_partitioned_level_timestamp
ON function_logs_partitioned(level, timestamp DESC);

-- Create monthly partitions
CREATE TABLE IF NOT EXISTS function_logs_y2025m01 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE IF NOT EXISTS function_logs_y2025m02 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

CREATE TABLE IF NOT EXISTS function_logs_y2025m03 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2025-03-01') TO ('2025-04-01');

CREATE TABLE IF NOT EXISTS function_logs_default PARTITION OF function_logs_partitioned
DEFAULT;

-- ============================================
-- Partition Maintenance Functions
-- ============================================

-- Function to create next month's partitions automatically
CREATE OR REPLACE FUNCTION create_next_month_partitions()
RETURNS void AS $$
DECLARE
    next_month_start DATE;
    next_month_end DATE;
    partition_name TEXT;
BEGIN
    -- Calculate next month
    next_month_start := date_trunc('month', CURRENT_DATE + INTERVAL '1 month');
    next_month_end := next_month_start + INTERVAL '1 month';

    -- Create partitions for each partitioned table
    FOREACH partition_name IN ARRAY ARRAY[
        'registry_function_executions_y',
        'performance_metrics_y',
        'system_health_checks_y',
        'database_metrics_y',
        'function_logs_y'
    ] LOOP
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %s%s%s PARTITION OF %s_partitioned FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            to_char(next_month_start, 'YYYYmMM'),
            '01',
            split_part(partition_name, '_y', 1),
            next_month_start,
            next_month_end
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to drop old partitions (older than 1 year)
CREATE OR REPLACE FUNCTION drop_old_partitions()
RETURNS void AS $$
DECLARE
    cutoff_date DATE := CURRENT_DATE - INTERVAL '1 year';
    partition_name TEXT;
    table_name TEXT;
BEGIN
    -- Drop old partitions for each partitioned table
    FOR table_name IN SELECT unnest(ARRAY[
        'registry_function_executions',
        'performance_metrics',
        'system_health_checks',
        'database_metrics',
        'function_logs'
    ]) LOOP
        FOR partition_name IN
            SELECT tablename FROM pg_tables
            WHERE tablename LIKE table_name || '_y%'
            AND tablename ~ ('_y\d{4}m\d{2}$')
            AND substring(tablename from char_length(table_name) - 6) < to_char(cutoff_date, 'YYMM')
        LOOP
            EXECUTE format('DROP TABLE IF EXISTS %I', partition_name);
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Comments for Documentation
-- ============================================

COMMENT ON TABLE registry_function_executions_partitioned IS 'Partitioned table for high-volume function execution data. Partitioned monthly by timestamp.';
COMMENT ON TABLE performance_metrics_partitioned IS 'Partitioned table for performance metrics. Partitioned monthly by timestamp.';
COMMENT ON TABLE system_health_checks_partitioned IS 'Partitioned table for system health check data. Partitioned monthly by checked_at.';
COMMENT ON TABLE database_metrics_partitioned IS 'Partitioned table for database performance metrics. Partitioned monthly by recorded_at.';
COMMENT ON TABLE function_logs_partitioned IS 'Partitioned table for function execution logs. Partitioned monthly by timestamp.';