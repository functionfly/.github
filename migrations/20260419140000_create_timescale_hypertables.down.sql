-- Migration: Rollback TimescaleDB Hypertables
-- Removes TimescaleDB-specific tables and restores original structure references

-- Drop continuous aggregates first
DROP MATERIALIZED VIEW IF EXISTS execution_hourly_summary CASCADE;
DROP MATERIALIZED VIEW IF EXISTS cost_daily_summary CASCADE;
DROP MATERIALIZED VIEW IF EXISTS cost_hourly_summary CASCADE;

-- Drop hypertables (this automatically drops chunks too)
DROP TABLE IF EXISTS cost_allocation_entries_ts CASCADE;
DROP TABLE IF EXISTS registry_function_executions_ts CASCADE;
DROP TABLE IF EXISTS health_checks_ts CASCADE;
DROP TABLE IF EXISTS routing_events_ts CASCADE;
DROP TABLE IF EXISTS performance_metrics_ts CASCADE;

-- Drop helper functions
DROP FUNCTION IF EXISTS migrate_to_timescale_tables(INTEGER, INTEGER);
DROP FUNCTION IF EXISTS set_tenant_retention_policy(UUID, INTEGER);

-- Drop tenant retention tracking
DROP TABLE IF EXISTS tenant_retention_policies CASCADE;
