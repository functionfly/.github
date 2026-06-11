-- Rollback: Remove daemon mode and signing key columns from agent_identities
ALTER TABLE agent_identities DROP COLUMN IF EXISTS signing_key_hash;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS swarm_topology;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS is_daemon_running;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS daemon_started_at;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS always_on_count;
ALTER TABLE agent_identities DROP COLUMN IF EXISTS daemon_execution_count;
