-- Migration: Remove R2 storage support columns from state fabric tables
-- Revert changes from 000250_state_fabric_r2_storage.up.sql

-- Remove R2 columns from state_fabric_events
ALTER TABLE state_fabric_events
    DROP COLUMN IF EXISTS r2_object_key,
    DROP COLUMN IF EXISTS r2_bucket,
    DROP COLUMN IF EXISTS batch_id,
    DROP COLUMN IF EXISTS is_archived,
    DROP COLUMN IF EXISTS archived_at;

-- Remove R2 columns from state_fabric_snapshots
ALTER TABLE state_fabric_snapshots
    DROP COLUMN IF EXISTS r2_object_key,
    DROP COLUMN IF EXISTS r2_bucket,
    DROP COLUMN IF EXISTS r2_content_hash;

-- Remove R2 columns from state_fabric_replays
ALTER TABLE state_fabric_replays
    DROP COLUMN IF EXISTS r2_object_key,
    DROP COLUMN IF EXISTS r2_bucket,
    DROP COLUMN IF EXISTS r2_content_hash;

-- Remove R2 columns from state_fabric_stores
ALTER TABLE state_fabric_stores
    DROP COLUMN IF EXISTS r2_memory_bucket,
    DROP COLUMN IF EXISTS r2_memory_enabled;

-- Remove R2 columns from state_snapshots
ALTER TABLE state_snapshots
    DROP COLUMN IF EXISTS r2_object_key,
    DROP COLUMN IF EXISTS r2_bucket,
    DROP COLUMN IF EXISTS r2_content_hash;

-- Remove R2 columns from state_events
ALTER TABLE state_events
    DROP COLUMN IF EXISTS r2_object_key,
    DROP COLUMN IF EXISTS r2_bucket,
    DROP COLUMN IF EXISTS batch_id,
    DROP COLUMN IF EXISTS is_archived,
    DROP COLUMN IF EXISTS archived_at;

-- Remove R2 columns from agent_memories
ALTER TABLE agent_memories
    DROP COLUMN IF EXISTS r2_object_key,
    DROP COLUMN IF EXISTS r2_bucket,
    DROP COLUMN IF EXISTS r2_content_hash,
    DROP COLUMN IF EXISTS is_offloaded,
    DROP COLUMN IF EXISTS offloaded_at;
