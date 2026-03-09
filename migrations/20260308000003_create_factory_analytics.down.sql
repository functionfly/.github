-- Rollback factory analytics tables
-- Migration: 20260308000003_create_factory_analytics.down.sql

-- Drop views first
DROP VIEW IF EXISTS factory_analytics_recent_runs;
DROP VIEW IF EXISTS factory_analytics_dashboard;

-- Drop functions
DROP FUNCTION IF EXISTS aggregate_factory_metrics(TEXT, TEXT, TIMESTAMPTZ);
DROP FUNCTION IF EXISTS cleanup_old_factory_metrics(INTEGER);

-- Drop tables
DROP TABLE IF EXISTS factory_analytics_aggregated;
DROP TABLE IF EXISTS factory_analytics_metrics;

-- Drop enum types (optional, may want to keep for other uses)
-- DROP TYPE IF EXISTS metric_type;
-- DROP TYPE IF EXISTS aggregation_period;
