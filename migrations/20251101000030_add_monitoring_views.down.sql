-- Remove database performance monitoring views and functions

-- Drop functions
DROP FUNCTION IF EXISTS get_db_recommendations();
DROP FUNCTION IF EXISTS collect_database_metrics();

-- Drop views
DROP VIEW IF EXISTS db_health_score;
DROP VIEW IF EXISTS replication_status;
DROP VIEW IF EXISTS tenant_usage_summary;
DROP VIEW IF EXISTS db_growth_trends;
DROP VIEW IF EXISTS db_lock_monitoring;
DROP VIEW IF EXISTS db_connection_stats;
DROP VIEW IF EXISTS db_query_performance;
DROP VIEW IF EXISTS db_unused_indexes;
DROP VIEW IF EXISTS db_index_usage;
DROP VIEW IF EXISTS db_table_sizes;