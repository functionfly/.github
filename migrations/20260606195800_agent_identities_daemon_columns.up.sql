-- Add missing daemon mode and signing key columns to agent_identities
-- These columns are expected by the Go model but were never added via migration

ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS signing_key_hash TEXT;
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS swarm_topology TEXT NOT NULL DEFAULT 'chain';
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS daemon_config JSONB NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS is_daemon_running BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS daemon_started_at TIMESTAMPTZ;
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS always_on_count INT NOT NULL DEFAULT 0;
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS daemon_execution_count BIGINT NOT NULL DEFAULT 0;
