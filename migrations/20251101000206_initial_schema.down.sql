-- Down migration for initial schema
-- Note: This is not a true reversal as it drops tables that may contain data.
-- Guard: this down migration raises an error so it never runs automatically.
-- To run after backing up data: set session variable then execute the DROP statements below manually.
DO $$
BEGIN
  IF current_setting('app.allow_000216_down', true) IS DISTINCT FROM 'true' THEN
    RAISE EXCEPTION 'Destructive down migration 000216 is disabled. Back up data first, then run: SET app.allow_000216_down = ''true''; and execute the DROP statements in this file manually.';
  END IF;
END $$;

-- Drop indexes (redundant with CASCADE but keeps downgrade explicit)
DROP INDEX IF EXISTS idx_routing_events_timestamp;
DROP INDEX IF EXISTS idx_routing_events_app_id;
DROP INDEX IF EXISTS idx_health_checks_backend_id;
DROP INDEX IF EXISTS idx_backends_app_id;

-- Drop tables in reverse dependency order (children first); CASCADE removes dependent objects
DROP TABLE IF EXISTS routing_events CASCADE;
DROP TABLE IF EXISTS circuit_state CASCADE;
DROP TABLE IF EXISTS health_checks CASCADE;
DROP TABLE IF EXISTS backends CASCADE;
DROP TABLE IF EXISTS apps CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;
DROP TABLE IF EXISTS schema_migrations CASCADE;
