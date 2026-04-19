-- Migration: Add R2 storage support columns to state fabric tables
-- This enables Cloudflare R2 storage for event logs, snapshots, memory blobs, and replay data

-- Add R2 storage columns to state_fabric_events table
ALTER TABLE state_fabric_events
    ADD COLUMN IF NOT EXISTS r2_object_key TEXT,
    ADD COLUMN IF NOT EXISTS r2_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS batch_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

-- Create index for archived events lookup
CREATE INDEX IF NOT EXISTS idx_state_fabric_events_archived ON state_fabric_events(is_archived, archived_at);
CREATE INDEX IF NOT EXISTS idx_state_fabric_events_batch ON state_fabric_events(batch_id) WHERE batch_id IS NOT NULL;

-- Add R2 storage columns to state_fabric_snapshots table
ALTER TABLE state_fabric_snapshots
    ADD COLUMN IF NOT EXISTS r2_object_key TEXT,
    ADD COLUMN IF NOT EXISTS r2_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS r2_content_hash VARCHAR(64);

-- Create index for R2 object lookups
CREATE INDEX IF NOT EXISTS idx_state_fabric_snapshots_r2 ON state_fabric_snapshots(r2_object_key) WHERE r2_object_key IS NOT NULL;

-- Add R2 storage columns to state_fabric_replays table
ALTER TABLE state_fabric_replays
    ADD COLUMN IF NOT EXISTS r2_object_key TEXT,
    ADD COLUMN IF NOT EXISTS r2_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS r2_content_hash VARCHAR(64);

-- Create index for R2 object lookups
CREATE INDEX IF NOT EXISTS idx_state_fabric_replays_r2 ON state_fabric_replays(r2_object_key) WHERE r2_object_key IS NOT NULL;

-- Add R2 memory storage columns to state_fabric_stores table
ALTER TABLE state_fabric_stores
    ADD COLUMN IF NOT EXISTS r2_memory_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS r2_memory_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- Create index for R2-enabled stores
CREATE INDEX IF NOT EXISTS idx_state_fabric_stores_r2_memory ON state_fabric_stores(r2_memory_enabled) WHERE r2_memory_enabled = TRUE;

-- Also update the underlying state tables (core state storage) for consistency
-- Add R2 columns to state_snapshots table
ALTER TABLE state_snapshots
    ADD COLUMN IF NOT EXISTS r2_object_key TEXT,
    ADD COLUMN IF NOT EXISTS r2_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS r2_content_hash VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_state_snapshots_r2 ON state_snapshots(r2_object_key) WHERE r2_object_key IS NOT NULL;

-- Add R2 columns to state_events table (for archival support)
ALTER TABLE state_events
    ADD COLUMN IF NOT EXISTS r2_object_key TEXT,
    ADD COLUMN IF NOT EXISTS r2_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS batch_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_state_events_archived ON state_events(is_archived, archived_at);
CREATE INDEX IF NOT EXISTS idx_state_events_batch ON state_events(batch_id) WHERE batch_id IS NOT NULL;

-- Add R2 columns to agent_memories table for memory blob storage
ALTER TABLE agent_memories
    ADD COLUMN IF NOT EXISTS r2_object_key TEXT,
    ADD COLUMN IF NOT EXISTS r2_bucket VARCHAR(255),
    ADD COLUMN IF NOT EXISTS r2_content_hash VARCHAR(64),
    ADD COLUMN IF NOT EXISTS is_offloaded BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS offloaded_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_agent_memories_offloaded ON agent_memories(is_offloaded, offloaded_at);
CREATE INDEX IF NOT EXISTS idx_agent_memories_r2 ON agent_memories(r2_object_key) WHERE r2_object_key IS NOT NULL;

-- Add comment to document the migration
COMMENT ON TABLE state_fabric_events IS 'State fabric events with optional R2 archival support';
COMMENT ON TABLE state_fabric_snapshots IS 'State fabric snapshots with optional R2 offloading';
COMMENT ON TABLE state_fabric_replays IS 'State fabric replays with R2 data storage';
COMMENT ON TABLE state_fabric_stores IS 'State fabric stores with R2 memory blob support';
