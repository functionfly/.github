-- Migration: 20260428233000_agent_lifecycle_management.up.sql
-- Adds lifecycle management fields to agent_identities table

-- Add lifecycle status and timestamps
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS lifecycle_status TEXT NOT NULL DEFAULT 'registered';
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS last_heartbeat_at TIMESTAMPTZ;
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMPTZ;
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS graceful_shutdown_at TIMESTAMPTZ;
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS forced_shutdown_at TIMESTAMPTZ;
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS orphan_detected_at TIMESTAMPTZ;

-- Add shutdown grace period configuration (seconds)
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS shutdown_grace_period_seconds INT NOT NULL DEFAULT 30;

-- Add persistent state snapshot (JSONB for crash recovery)
ALTER TABLE agent_identities ADD COLUMN IF NOT EXISTS state_snapshot JSONB DEFAULT '{}';

-- Add lifecycle event log
CREATE TABLE IF NOT EXISTS agent_lifecycle_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    event_type TEXT NOT NULL, -- registered, heartbeat, graceful_shutdown_start, graceful_shutdown_complete, forced_shutdown, orphan_detected, crash_recovery
    event_data JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_lifecycle_events_agent_id ON agent_lifecycle_events(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_lifecycle_events_type ON agent_lifecycle_events(event_type);
CREATE INDEX IF NOT EXISTS idx_agent_lifecycle_events_created_at ON agent_lifecycle_events(created_at);

-- Add index for orphan detection (stale agents)
CREATE INDEX IF NOT EXISTS idx_agent_identities_heartbeat ON agent_identities(last_heartbeat_at) WHERE last_heartbeat_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_agent_identities_lifecycle_status ON agent_identities(lifecycle_status);