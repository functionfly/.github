-- Migration: 20260428233000_agent_lifecycle_management.down.sql
-- Removes lifecycle management fields from agent_identities table

ALTER TABLE agent_identities DROP COLUMN IF EXISTS lifecycle_status;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS last_heartbeat_at;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS last_active_at;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS graceful_shutdown_at;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS forced_shutdown_at;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS orphan_detected_at;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS shutdown_grace_period_seconds;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS state_snapshot;

DROP TABLE IF EXISTS agent_lifecycle_events;