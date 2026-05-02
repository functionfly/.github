-- Migration: Remove R2 storage columns from state fabric and related tables
-- Reverses: 20260412180100_state_fabric_r2_storage

-- Remove from state_fabric_events
DROP INDEX IF EXISTS idx_state_fabric_events_archived;
DROP INDEX IF EXISTS idx_state_fabric_events_batch;
ALTER TABLE state_fabric_events DROP COLUMN IF EXISTS r2_object_key;
ALTER TABLE state_fabric_events DROP COLUMN IF EXISTS r2_bucket;
ALTER TABLE state_fabric_events DROP COLUMN IF EXISTS batch_id;
ALTER TABLE state_fabric_events DROP COLUMN IF EXISTS is_archived;
ALTER TABLE state_fabric_events DROP COLUMN IF EXISTS archived_at;

-- Remove from state_fabric_snapshots
DROP INDEX IF EXISTS idx_state_fabric_snapshots_r2;
ALTER TABLE state_fabric_snapshots DROP COLUMN IF EXISTS r2_object_key;
ALTER TABLE state_fabric_snapshots DROP COLUMN IF EXISTS r2_bucket;
ALTER TABLE state_fabric_snapshots DROP COLUMN IF EXISTS r2_content_hash;

-- Remove from state_fabric_replays
DROP INDEX IF EXISTS idx_state_fabric_replays_r2;
ALTER TABLE state_fabric_replays DROP COLUMN IF EXISTS r2_object_key;
ALTER TABLE state_fabric_replays DROP COLUMN IF EXISTS r2_bucket;
ALTER TABLE state_fabric_replays DROP COLUMN IF EXISTS r2_content_hash;

-- Remove from state_fabric_stores
DROP INDEX IF EXISTS idx_state_fabric_stores_r2_memory;
ALTER TABLE state_fabric_stores DROP COLUMN IF EXISTS r2_memory_bucket;
ALTER TABLE state_fabric_stores DROP COLUMN IF EXISTS r2_memory_enabled;

-- Remove from state_snapshots
DROP INDEX IF EXISTS idx_state_snapshots_r2;
ALTER TABLE state_snapshots DROP COLUMN IF EXISTS r2_object_key;
ALTER TABLE state_snapshots DROP COLUMN IF EXISTS r2_bucket;
ALTER TABLE state_snapshots DROP COLUMN IF EXISTS r2_content_hash;

-- Remove from state_events
DROP INDEX IF EXISTS idx_state_events_archived;
DROP INDEX IF EXISTS idx_state_events_batch;
ALTER TABLE state_events DROP COLUMN IF EXISTS r2_object_key;
ALTER TABLE state_events DROP COLUMN IF EXISTS r2_bucket;
ALTER TABLE state_events DROP COLUMN IF EXISTS batch_id;
ALTER TABLE state_events DROP COLUMN IF EXISTS is_archived;
ALTER TABLE state_events DROP COLUMN IF EXISTS archived_at;

-- Remove from agent_memories
DROP INDEX IF EXISTS idx_agent_memories_offloaded;
DROP INDEX IF EXISTS idx_agent_memories_r2;
ALTER TABLE agent_memories DROP COLUMN IF EXISTS r2_object_key;
ALTER TABLE agent_memories DROP COLUMN IF EXISTS r2_bucket;
ALTER TABLE agent_memories DROP COLUMN IF EXISTS r2_content_hash;
ALTER TABLE agent_memories DROP COLUMN IF EXISTS is_offloaded;
ALTER TABLE agent_memories DROP COLUMN IF EXISTS offloaded_at;