-- Dynamic partition creation for 2026 and beyond
-- This migration adds partitions for the current year and next 3 months
-- and replaces the static partition creation function with a dynamic one

-- ============================================
-- Add 2026 Partitions for Registry Function Executions
-- ============================================

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m01 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m02 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m03 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m04 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m05 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m06 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m07 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m08 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m09 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m10 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m11 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2026m12 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

-- Add 2027 partitions (first 3 months for runway)
CREATE TABLE IF NOT EXISTS registry_function_executions_y2027m01 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2027m02 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');

CREATE TABLE IF NOT EXISTS registry_function_executions_y2027m03 PARTITION OF registry_function_executions_partitioned
FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');

-- ============================================
-- Add 2026 Partitions for Performance Metrics
-- ============================================

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m01 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m02 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m03 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m04 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m05 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m06 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m07 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m08 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m09 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m10 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m11 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2026m12 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2027m01 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2027m02 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');

CREATE TABLE IF NOT EXISTS performance_metrics_y2027m03 PARTITION OF performance_metrics_partitioned
FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');

-- ============================================
-- Add 2026 Partitions for System Health Checks
-- ============================================

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m01 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m02 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m03 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m04 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m05 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m06 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m07 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m08 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m09 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m10 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m11 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2026m12 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2027m01 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2027m02 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');

CREATE TABLE IF NOT EXISTS system_health_checks_y2027m03 PARTITION OF system_health_checks_partitioned
FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');

-- ============================================
-- Add 2026 Partitions for Database Metrics
-- ============================================

CREATE TABLE IF NOT EXISTS database_metrics_y2026m01 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m02 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m03 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m04 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m05 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m06 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m07 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m08 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m09 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m10 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m11 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2026m12 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2027m01 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2027m02 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');

CREATE TABLE IF NOT EXISTS database_metrics_y2027m03 PARTITION OF database_metrics_partitioned
FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');

-- ============================================
-- Add 2026 Partitions for Function Logs
-- ============================================

CREATE TABLE IF NOT EXISTS function_logs_y2026m01 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m02 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m03 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m04 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m05 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m06 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m07 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m08 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m09 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m10 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m11 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');

CREATE TABLE IF NOT EXISTS function_logs_y2026m12 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

CREATE TABLE IF NOT EXISTS function_logs_y2027m01 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2027-01-01') TO ('2027-02-01');

CREATE TABLE IF NOT EXISTS function_logs_y2027m02 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2027-02-01') TO ('2027-03-01');

CREATE TABLE IF NOT EXISTS function_logs_y2027m03 PARTITION OF function_logs_partitioned
FOR VALUES FROM ('2027-03-01') TO ('2027-04-01');

-- ============================================
-- Replace Partition Creation Function with Dynamic Version
-- ============================================

-- Drop the old function
DROP FUNCTION IF EXISTS create_next_month_partitions();

-- Create new dynamic function that generates partitions for current + 3 months
CREATE OR REPLACE FUNCTION create_next_month_partitions()
RETURNS void AS $$
DECLARE
    current_date_val DATE := CURRENT_DATE;
    i INTEGER;
    partition_year TEXT;
    partition_month TEXT;
    partition_name TEXT;
    table_prefix TEXT;
    partition_start DATE;
    partition_end DATE;
BEGIN
    -- Create partitions for current month + next 3 months
    FOR i IN 0..3 LOOP
        partition_start := date_trunc('month', current_date_val + (i || ' months')::INTERVAL);
        partition_end := partition_start + INTERVAL '1 month';
        partition_year := to_char(partition_start, 'YYYY');
        partition_month := to_char(partition_start, 'MM');

        -- Create partitions for each partitioned table
        FOREACH table_prefix IN ARRAY ARRAY[
            'registry_function_executions',
            'performance_metrics',
            'system_health_checks',
            'database_metrics',
            'function_logs'
        ] LOOP
            partition_name := table_prefix || '_y' || partition_year || 'm' || partition_month;
            EXECUTE format(
                'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I_partitioned FOR VALUES FROM (%L) TO (%L)',
                partition_name,
                table_prefix,
                partition_start,
                partition_end
            );
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Add Missing Indexes on routing_events
-- ============================================

-- Index for backend_id queries (health monitoring, failure analysis)
CREATE INDEX IF NOT EXISTS idx_routing_events_backend_id_timestamp
ON routing_events(backend_id, timestamp DESC);

-- Index for outcome queries (failure rate calculations, success rate metrics)
CREATE INDEX IF NOT EXISTS idx_routing_events_outcome_timestamp
ON routing_events(outcome, timestamp DESC);

-- Composite index for common query pattern: app + backend + time range
CREATE INDEX IF NOT EXISTS idx_routing_events_app_backend_timestamp
ON routing_events(app_id, backend_id, timestamp DESC);

-- ============================================
-- Add CHECK Constraints on Monetary Columns
-- ============================================

-- Add CHECK constraints to cost_allocation_entries if not already present
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_costs;
ALTER TABLE cost_allocation_entries ADD CONSTRAINT chk_cost_allocation_entries_positive_costs
    CHECK (execution_cost_cents >= 0);
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_compute;
ALTER TABLE cost_allocation_entries ADD CONSTRAINT chk_cost_allocation_entries_positive_compute
    CHECK (compute_cost_cents >= 0);
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_platform_fee;
ALTER TABLE cost_allocation_entries ADD CONSTRAINT chk_cost_allocation_entries_positive_platform_fee
    CHECK (platform_fee_cents >= 0);
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_data_transfer;
ALTER TABLE cost_allocation_entries ADD CONSTRAINT chk_cost_allocation_entries_positive_data_transfer
    CHECK (data_transfer_cents >= 0);
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_total;
ALTER TABLE cost_allocation_entries ADD CONSTRAINT chk_cost_allocation_entries_positive_total
    CHECK (total_cost_cents >= 0);
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_non_negative_duration;
ALTER TABLE cost_allocation_entries ADD CONSTRAINT chk_cost_allocation_entries_non_negative_duration
    CHECK (duration_ms >= 0);
