-- Down migration: initial schema (routing/load-balancer tables)
--
-- DESTRUCTIVE: Drops tables and all data. Use only in dev, CI, or after
-- backup. In production, prefer forward-only migrations or restore from
-- backup instead of running down migrations.
--
-- Drop order: child tables first to satisfy foreign keys (no CASCADE).

-- Indexes (optional; Postgres drops them with the table, but explicit for clarity)
DROP INDEX IF EXISTS idx_routing_events_timestamp;
DROP INDEX IF EXISTS idx_routing_events_app_id;
DROP INDEX IF EXISTS idx_health_checks_backend_id;
DROP INDEX IF EXISTS idx_backends_app_id;

-- Child tables first (depend on backends / apps / tenants)
DROP TABLE IF EXISTS routing_events;
DROP TABLE IF EXISTS circuit_state;
DROP TABLE IF EXISTS health_checks;
DROP TABLE IF EXISTS backends;
DROP TABLE IF EXISTS apps;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;

-- Migration tracking (golang-migrate); omit if your runner manages this
DROP TABLE IF EXISTS schema_migrations;
