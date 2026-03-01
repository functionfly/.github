-- Drop table partitioning

-- Drop partitioned tables and their partitions
DROP TABLE IF EXISTS registry_function_executions_partitioned CASCADE;
DROP TABLE IF EXISTS performance_metrics_partitioned CASCADE;
DROP TABLE IF EXISTS system_health_checks_partitioned CASCADE;
DROP TABLE IF EXISTS database_metrics_partitioned CASCADE;
DROP TABLE IF EXISTS function_logs_partitioned CASCADE;

-- Drop partition management functions
DROP FUNCTION IF EXISTS create_next_month_partitions();
DROP FUNCTION IF EXISTS drop_old_partitions();