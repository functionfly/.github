-- Down migration for initial schema
-- Note: This is not a true reversal as it drops tables that may contain data
-- In production, you would want more careful down migrations

DROP INDEX IF EXISTS idx_routing_events_timestamp;
DROP INDEX IF EXISTS idx_routing_events_app_id;
DROP INDEX IF EXISTS idx_health_checks_backend_id;
DROP INDEX IF EXISTS idx_backends_app_id;

DROP TABLE IF EXISTS routing_events;
DROP TABLE IF EXISTS circuit_state;
DROP TABLE IF EXISTS health_checks;
DROP TABLE IF EXISTS backends;
DROP TABLE IF EXISTS apps;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS schema_migrations;