-- Migration: Rollback Fixed Native Partitioning
-- Drops all partitioned tables and functions

-- Drop partition management functions
DROP FUNCTION IF EXISTS auto_partition_maintenance();
DROP FUNCTION IF EXISTS migrate_to_partitioned_table(TEXT, TEXT, DATE, DATE, INTEGER);
DROP FUNCTION IF EXISTS drop_old_partitions_native(TEXT, INTEGER);
DROP FUNCTION IF EXISTS create_next_month_partitions_native();

-- Drop partitioned tables (this automatically drops partitions)
DROP TABLE IF EXISTS function_logs_partitioned CASCADE;
DROP TABLE IF EXISTS performance_metrics_partitioned CASCADE;
DROP TABLE IF EXISTS health_checks_partitioned CASCADE;
DROP TABLE IF EXISTS routing_events_partitioned CASCADE;
DROP TABLE IF EXISTS registry_function_executions_partitioned CASCADE;
DROP TABLE IF EXISTS cost_allocation_entries_partitioned CASCADE;
