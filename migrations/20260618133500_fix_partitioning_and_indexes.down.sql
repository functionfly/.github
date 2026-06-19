-- Down migration for 20260618133500_fix_partitioning_and_indexes
-- Reverses: partition creation, dynamic partition function, routing_events indexes, and cost_allocation CHECK constraints

-- ============================================
-- Drop CHECK Constraints on cost_allocation_entries
-- ============================================

ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_costs;
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_compute;
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_platform_fee;
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_data_transfer;
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_positive_total;
ALTER TABLE cost_allocation_entries DROP CONSTRAINT IF EXISTS chk_cost_allocation_entries_non_negative_duration;

-- ============================================
-- Drop Indexes on routing_events
-- ============================================

DROP INDEX IF EXISTS idx_routing_events_backend_id_timestamp;
DROP INDEX IF EXISTS idx_routing_events_outcome_timestamp;
DROP INDEX IF EXISTS idx_routing_events_app_backend_timestamp;

-- ============================================
-- Restore Original Partition Creation Function
-- ============================================

DROP FUNCTION IF EXISTS create_next_month_partitions();

-- Recreate a basic partition creation function (creates next month's partitions only)
CREATE OR REPLACE FUNCTION create_next_month_partitions()
RETURNS void AS $$
DECLARE
    current_date_val DATE := CURRENT_DATE;
    next_month_start DATE;
    next_month_end DATE;
    partition_year TEXT;
    partition_month TEXT;
    partition_name TEXT;
    table_prefix TEXT;
BEGIN
    -- Calculate next month boundaries
    next_month_start := date_trunc('month', current_date_val) + INTERVAL '1 month';
    next_month_end := next_month_start + INTERVAL '1 month';
    partition_year := to_char(next_month_start, 'YYYY');
    partition_month := to_char(next_month_start, 'MM');

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
            next_month_start,
            next_month_end
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Drop 2027 Partitions (only if empty - partitions persist data)
-- Note: In production, ensure these partitions are empty before dropping
-- The IF EXISTS prevents errors if partitions don't exist
-- ============================================

-- Registry Function Executions 2027
DROP TABLE IF EXISTS registry_function_executions_y2027m03;
DROP TABLE IF EXISTS registry_function_executions_y2027m02;
DROP TABLE IF EXISTS registry_function_executions_y2027m01;

-- Performance Metrics 2027
DROP TABLE IF EXISTS performance_metrics_y2027m03;
DROP TABLE IF EXISTS performance_metrics_y2027m02;
DROP TABLE IF EXISTS performance_metrics_y2027m01;

-- System Health Checks 2027
DROP TABLE IF EXISTS system_health_checks_y2027m03;
DROP TABLE IF EXISTS system_health_checks_y2027m02;
DROP TABLE IF EXISTS system_health_checks_y2027m01;

-- Database Metrics 2027
DROP TABLE IF EXISTS database_metrics_y2027m03;
DROP TABLE IF EXISTS database_metrics_y2027m02;
DROP TABLE IF EXISTS database_metrics_y2027m01;

-- Function Logs 2027
DROP TABLE IF EXISTS function_logs_y2027m03;
DROP TABLE IF EXISTS function_logs_y2027m02;
DROP TABLE IF EXISTS function_logs_y2027m01;

-- ============================================
-- Note: 2026 partitions are NOT dropped because they contain current-year data
-- If rollback to before June 2026 is truly needed, manually drop 2026 partitions
-- using: DROP TABLE IF EXISTS registry_function_executions_y2026mXX;
-- ============================================
