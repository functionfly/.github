-- Migration: Fixed Native PostgreSQL Partitioning
-- Fixes unique constraint issues for high-volume time-series tables
-- Uses UUID + timestamp composite keys to satisfy partitioning requirements
-- Created: 2026-04-19

-- ============================================
-- 1. Cost Allocation Entries Partitioned Table
-- ============================================

-- Drop existing partitioned tables if they exist (clean slate)
DROP TABLE IF EXISTS cost_allocation_entries_partitioned CASCADE;

-- Create partitioned table with proper unique constraint including partition key
CREATE TABLE cost_allocation_entries_partitioned (
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
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    tags JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Primary key must include partition key (timestamp)
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create default partition for data outside defined ranges
CREATE TABLE cost_allocation_entries_default PARTITION OF cost_allocation_entries_partitioned
DEFAULT;

-- Create monthly partitions (current year + next year)
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
    year_month TEXT;
BEGIN
    -- Create partitions for current month through 12 months ahead
    FOR i IN 0..12 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        year_month := to_char(start_date, 'YYYY_MM');
        partition_name := 'cost_allocation_entries_y' || year_month;
        
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF cost_allocation_entries_partitioned 
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            start_date,
            end_date
        );
    END LOOP;
END $$;

-- Indexes on partitioned table (inherited by partitions)
CREATE INDEX idx_cost_alloc_part_tenant_time 
ON cost_allocation_entries_partitioned (tenant_id, timestamp DESC);

CREATE INDEX idx_cost_alloc_part_function_time 
ON cost_allocation_entries_partitioned (function_id, timestamp DESC);

CREATE INDEX idx_cost_alloc_part_outcome_time 
ON cost_allocation_entries_partitioned (execution_outcome, timestamp DESC);

CREATE INDEX idx_cost_alloc_part_region_time 
ON cost_allocation_entries_partitioned (region, timestamp DESC) 
WHERE region IS NOT NULL;

-- GIN index for tags JSONB
CREATE INDEX idx_cost_alloc_part_tags 
ON cost_allocation_entries_partitioned USING GIN (tags);

-- Covering index for billing reports (avoid heap lookups)
CREATE INDEX idx_cost_alloc_part_tenant_covering 
ON cost_allocation_entries_partitioned (tenant_id, timestamp DESC)
INCLUDE (function_id, total_cost_cents, execution_outcome, cached);

-- ============================================
-- 2. Registry Function Executions Partitioned
-- ============================================

DROP TABLE IF EXISTS registry_function_executions_partitioned CASCADE;

CREATE TABLE registry_function_executions_partitioned (
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
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    verification_status TEXT,
    verification_error TEXT,
    replayed_duration_ms INTEGER,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE TABLE registry_function_executions_default PARTITION OF registry_function_executions_partitioned
DEFAULT;

-- Create monthly partitions
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
    year_month TEXT;
BEGIN
    FOR i IN 0..12 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        year_month := to_char(start_date, 'YYYY_MM');
        partition_name := 'registry_exec_y' || year_month;
        
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF registry_function_executions_partitioned 
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            start_date,
            end_date
        );
    END LOOP;
END $$;

-- Indexes
CREATE INDEX idx_registry_exec_part_function_time 
ON registry_function_executions_partitioned (function_id, timestamp DESC);

CREATE INDEX idx_registry_exec_part_tenant_time 
ON registry_function_executions_partitioned (tenant_id, timestamp DESC) 
WHERE tenant_id IS NOT NULL;

CREATE INDEX idx_registry_exec_part_outcome_time 
ON registry_function_executions_partitioned (outcome, timestamp DESC);

CREATE INDEX idx_registry_exec_part_verification 
ON registry_function_executions_partitioned (verification_status, timestamp DESC) 
WHERE verification_status IS NOT NULL;

-- Covering index for dashboard queries
CREATE INDEX idx_registry_exec_part_tenant_covering 
ON registry_function_executions_partitioned (tenant_id, timestamp DESC)
INCLUDE (function_id, outcome, duration_ms, cached)
WHERE tenant_id IS NOT NULL;

-- ============================================
-- 3. Routing Events Partitioned
-- ============================================

DROP TABLE IF EXISTS routing_events_partitioned CASCADE;

CREATE TABLE routing_events_partitioned (
    id UUID NOT NULL,
    app_id UUID NOT NULL,
    backend_id UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    latency_ms INTEGER,
    outcome VARCHAR(20) NOT NULL,
    request_id VARCHAR(255),
    region VARCHAR(50),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE TABLE routing_events_default PARTITION OF routing_events_partitioned
DEFAULT;

-- Create monthly partitions
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
    year_month TEXT;
BEGIN
    FOR i IN 0..12 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        year_month := to_char(start_date, 'YYYY_MM');
        partition_name := 'routing_events_y' || year_month;
        
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF routing_events_partitioned 
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            start_date,
            end_date
        );
    END LOOP;
END $$;

-- Indexes
CREATE INDEX idx_routing_part_app_time 
ON routing_events_partitioned (app_id, timestamp DESC);

CREATE INDEX idx_routing_part_backend_time 
ON routing_events_partitioned (backend_id, timestamp DESC);

CREATE INDEX idx_routing_part_outcome_time 
ON routing_events_partitioned (outcome, timestamp DESC);

CREATE INDEX idx_routing_part_region_time 
ON routing_events_partitioned (region, timestamp DESC) 
WHERE region IS NOT NULL;

-- ============================================
-- 4. Health Checks Partitioned
-- ============================================

DROP TABLE IF EXISTS health_checks_partitioned CASCADE;

CREATE TABLE health_checks_partitioned (
    id UUID NOT NULL,
    backend_id UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    ok BOOLEAN NOT NULL,
    status_code INTEGER,
    latency_ms INTEGER,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE TABLE health_checks_default PARTITION OF health_checks_partitioned
DEFAULT;

-- Create monthly partitions
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
    year_month TEXT;
BEGIN
    FOR i IN 0..12 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        year_month := to_char(start_date, 'YYYY_MM');
        partition_name := 'health_checks_y' || year_month;
        
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF health_checks_partitioned 
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            start_date,
            end_date
        );
    END LOOP;
END $$;

-- Indexes
CREATE INDEX idx_health_part_backend_time 
ON health_checks_partitioned (backend_id, timestamp DESC);

CREATE INDEX idx_health_part_ok_time 
ON health_checks_partitioned (ok, timestamp DESC) 
WHERE NOT ok;

CREATE INDEX idx_health_part_status_time 
ON health_checks_partitioned (status_code, timestamp DESC) 
WHERE status_code IS NOT NULL;

-- ============================================
-- 5. Performance Metrics Partitioned
-- ============================================

DROP TABLE IF EXISTS performance_metrics_partitioned CASCADE;

CREATE TABLE performance_metrics_partitioned (
    id UUID NOT NULL,
    metric_type VARCHAR(100) NOT NULL,
    tenant_id UUID,
    app_id UUID,
    backend_id UUID,
    value DECIMAL(20,6) NOT NULL,
    unit VARCHAR(20),
    labels JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE TABLE performance_metrics_default PARTITION OF performance_metrics_partitioned
DEFAULT;

-- Create monthly partitions
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
    year_month TEXT;
BEGIN
    FOR i IN 0..12 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        year_month := to_char(start_date, 'YYYY_MM');
        partition_name := 'perf_metrics_y' || year_month;
        
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF performance_metrics_partitioned 
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            start_date,
            end_date
        );
    END LOOP;
END $$;

-- Indexes
CREATE INDEX idx_perf_part_type_time 
ON performance_metrics_partitioned (metric_type, timestamp DESC);

CREATE INDEX idx_perf_part_tenant_time 
ON performance_metrics_partitioned (tenant_id, timestamp DESC) 
WHERE tenant_id IS NOT NULL;

CREATE INDEX idx_perf_part_app_time 
ON performance_metrics_partitioned (app_id, timestamp DESC) 
WHERE app_id IS NOT NULL;

CREATE INDEX idx_perf_part_labels 
ON performance_metrics_partitioned USING GIN (labels);

-- ============================================
-- 6. Function Logs Partitioned
-- ============================================

DROP TABLE IF EXISTS function_logs_partitioned CASCADE;

CREATE TABLE function_logs_partitioned (
    id UUID NOT NULL,
    function_id UUID,
    deployment_id UUID,
    level VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE TABLE function_logs_default PARTITION OF function_logs_partitioned
DEFAULT;

-- Create monthly partitions
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
    year_month TEXT;
BEGIN
    FOR i IN 0..12 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        year_month := to_char(start_date, 'YYYY_MM');
        partition_name := 'function_logs_y' || year_month;
        
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF function_logs_partitioned 
             FOR VALUES FROM (%L) TO (%L)',
            partition_name,
            start_date,
            end_date
        );
    END LOOP;
END $$;

-- Indexes
CREATE INDEX idx_logs_part_function_time 
ON function_logs_partitioned (function_id, timestamp DESC) 
WHERE function_id IS NOT NULL;

CREATE INDEX idx_logs_part_level_time 
ON function_logs_partitioned (level, timestamp DESC);

CREATE INDEX idx_logs_part_deployment_time 
ON function_logs_partitioned (deployment_id, timestamp DESC) 
WHERE deployment_id IS NOT NULL;

-- GIN index for log metadata search
CREATE INDEX idx_logs_part_metadata 
ON function_logs_partitioned USING GIN (metadata);

-- ============================================
-- 7. Partition Management Functions
-- ============================================

-- Function to create next month's partitions for all partitioned tables
CREATE OR REPLACE FUNCTION create_next_month_partitions_native()
RETURNS TABLE (table_name TEXT, partition_name TEXT, created BOOLEAN) AS $$
DECLARE
    next_month_start DATE;
    next_month_end DATE;
    ym TEXT;
    tbl_name TEXT;
    part_name TEXT;
    tables TEXT[] := ARRAY[
        'cost_allocation_entries',
        'registry_function_executions', 
        'routing_events',
        'health_checks',
        'performance_metrics',
        'function_logs'
    ];
BEGIN
    next_month_start := date_trunc('month', CURRENT_DATE + INTERVAL '1 month');
    next_month_end := next_month_start + INTERVAL '1 month';
    ym := to_char(next_month_start, 'YYYY_MM');
    
    FOREACH tbl_name IN ARRAY tables
    LOOP
        -- Different naming conventions per table
        CASE tbl_name
            WHEN 'cost_allocation_entries' THEN part_name := 'cost_allocation_entries_y' || ym;
            WHEN 'registry_function_executions' THEN part_name := 'registry_exec_y' || ym;
            WHEN 'routing_events' THEN part_name := 'routing_events_y' || ym;
            WHEN 'health_checks' THEN part_name := 'health_checks_y' || ym;
            WHEN 'performance_metrics' THEN part_name := 'perf_metrics_y' || ym;
            WHEN 'function_logs' THEN part_name := 'function_logs_y' || ym;
        END CASE;
        
        BEGIN
            EXECUTE format(
                'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I_partitioned 
                 FOR VALUES FROM (%L) TO (%L)',
                part_name,
                tbl_name,
                next_month_start,
                next_month_end
            );
            RETURN QUERY SELECT tbl_name, part_name, TRUE;
        EXCEPTION WHEN duplicate_table THEN
            RETURN QUERY SELECT tbl_name, part_name, FALSE;
        END;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to drop old partitions based on retention policy
CREATE OR REPLACE FUNCTION drop_old_partitions_native(
    p_table_name TEXT,
    p_retention_days INTEGER DEFAULT 90
)
RETURNS TABLE (dropped_partition TEXT, rows_affected BIGINT) AS $$
DECLARE
    cutoff_date DATE := CURRENT_DATE - (p_retention_days || ' days')::INTERVAL;
    partition_record RECORD;
    partition_date DATE;
    rows_in_partition BIGINT;
BEGIN
    FOR partition_record IN 
        SELECT inhrelid::regclass::text as partition_name
        FROM pg_inherits 
        WHERE inhparent = (p_table_name || '_partitioned')::regclass
    LOOP
        -- Extract date from partition name (assumes _yYYYY_MM suffix)
        IF partition_record.partition_name ~ '_y\d{4}_\d{2}$' THEN
            partition_date := to_date(
                substring(partition_record.partition_name from '_y(\d{4}_\d{2})$'), 
                'YYYY_MM'
            );
            
            IF partition_date < cutoff_date THEN
                -- Count rows before dropping (for audit)
                EXECUTE format('SELECT COUNT(*) FROM %I', partition_record.partition_name)
                INTO rows_in_partition;
                
                -- Log to retention audit
                INSERT INTO retention_audit_log (
                    table_name, partition_name, cutoff_date, 
                    rows_deleted, deleted_at, triggered_by
                ) VALUES (
                    p_table_name, partition_record.partition_name, cutoff_date,
                    rows_in_partition, NOW(), 'partition_maintenance'
                );
                
                -- Drop the partition
                EXECUTE format('DROP TABLE IF EXISTS %I', partition_record.partition_name);
                
                RETURN QUERY SELECT partition_record.partition_name, rows_in_partition;
            END IF;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to migrate data from old table to partitioned (batch insert)
CREATE OR REPLACE FUNCTION migrate_to_partitioned_table(
    p_source_table TEXT,
    p_target_table TEXT,
    p_start_date DATE,
    p_end_date DATE,
    p_batch_size INTEGER DEFAULT 10000
)
RETURNS TABLE (batches_processed INTEGER, total_rows_migrated BIGINT) AS $$
DECLARE
    v_batch_count INTEGER := 0;
    v_total_rows BIGINT := 0;
    v_batch_rows INTEGER;
BEGIN
    LOOP
        -- Insert batch
        EXECUTE format(
            'WITH batch AS (
                SELECT * FROM %I 
                WHERE timestamp >= %L AND timestamp < %L
                AND id NOT IN (SELECT id FROM %I_partitioned WHERE timestamp >= %L AND timestamp < %L)
                LIMIT %s
                FOR UPDATE SKIP LOCKED
            )
            INSERT INTO %I_partitioned 
            SELECT * FROM batch',
            p_source_table, p_start_date, p_end_date,
            p_target_table, p_start_date, p_end_date,
            p_batch_size,
            p_target_table
        );
        
        GET DIAGNOSTICS v_batch_rows = ROW_COUNT;
        
        v_batch_count := v_batch_count + 1;
        v_total_rows := v_total_rows + v_batch_rows;
        
        EXIT WHEN v_batch_rows = 0;
        
        -- Commit and sleep briefly to reduce lock contention
        COMMIT;
        PERFORM pg_sleep(0.1);
    END LOOP;
    
    RETURN QUERY SELECT v_batch_count, v_total_rows;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 8. Auto-partition Trigger (run via cron/pg_cron)
-- ============================================

CREATE OR REPLACE FUNCTION auto_partition_maintenance()
RETURNS VOID AS $$
BEGIN
    -- Create next month partitions
    PERFORM create_next_month_partitions_native();
    
    -- Drop old partitions based on retention settings
    PERFORM drop_old_partitions_native('cost_allocation_entries', 90);
    PERFORM drop_old_partitions_native('registry_function_executions', 30);
    PERFORM drop_old_partitions_native('routing_events', 30);
    PERFORM drop_old_partitions_native('health_checks', 90);
    PERFORM drop_old_partitions_native('performance_metrics', 30);
    PERFORM drop_old_partitions_native('function_logs', 30);
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION auto_partition_maintenance() IS 
'Run monthly via cron/pg_cron: SELECT auto_partition_maintenance();';

-- ============================================
-- 9. Comments for Documentation
-- ============================================

COMMENT ON TABLE cost_allocation_entries_partitioned IS 
'Native PostgreSQL partitioned table for cost allocation data. Monthly partitions, 90-day retention.';

COMMENT ON TABLE registry_function_executions_partitioned IS 
'Native PostgreSQL partitioned table for function executions. Monthly partitions, 30-day retention.';

COMMENT ON TABLE routing_events_partitioned IS 
'Native PostgreSQL partitioned table for routing events. Monthly partitions, 30-day retention.';

COMMENT ON TABLE health_checks_partitioned IS 
'Native PostgreSQL partitioned table for health checks. Monthly partitions, 90-day retention.';

COMMENT ON TABLE performance_metrics_partitioned IS 
'Native PostgreSQL partitioned table for performance metrics. Monthly partitions, 30-day retention.';

COMMENT ON TABLE function_logs_partitioned IS 
'Native PostgreSQL partitioned table for function logs. Monthly partitions, 30-day retention.';
