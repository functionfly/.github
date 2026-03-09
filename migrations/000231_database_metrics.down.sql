-- Drop database metrics table and related objects

DROP TABLE IF EXISTS database_metrics;
DROP FUNCTION IF EXISTS cleanup_old_database_metrics();