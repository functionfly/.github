-- Migration: Rollback Monitoring and Operational Functions

-- Drop functions (dependent first)
DROP FUNCTION IF EXISTS log_maintenance_job(TEXT, TEXT, INTERVAL, TEXT);
DROP FUNCTION IF EXISTS maintenance_reindex_bloated();
DROP FUNCTION IF EXISTS maintenance_vacuum_all();
DROP FUNCTION IF EXISTS detailed_health_report();
DROP FUNCTION IF EXISTS quick_health_check();
DROP FUNCTION IF EXISTS check_replication_lag();
DROP FUNCTION IF EXISTS get_large_tables(INTEGER);
DROP FUNCTION IF EXISTS get_database_size();
DROP FUNCTION IF EXISTS vacuum_bloated_tables(INTEGER, BOOLEAN);
DROP FUNCTION IF EXISTS estimate_table_bloat(TEXT);
DROP FUNCTION IF EXISTS terminate_long_query(INTEGER, BOOLEAN);

-- Drop views
DROP VIEW IF EXISTS replication_status;
DROP VIEW IF EXISTS missing_index_candidates;
DROP VIEW IF EXISTS slow_queries;
DROP VIEW IF EXISTS autovacuum_status;
DROP VIEW IF EXISTS index_health;
DROP VIEW IF EXISTS table_bloat;
DROP VIEW IF EXISTS lock_conflicts;
DROP VIEW IF EXISTS active_connections;
DROP VIEW IF EXISTS database_size_breakdown;

-- Drop table
DROP TABLE IF EXISTS db_maintenance_jobs;
