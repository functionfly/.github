-- Migration: Add target_memory_id to memory_shares for tracking shared memory instances
-- Created: 2026-04-11
-- Purpose: Track the actual memory record created in target team for proper revocation

ALTER TABLE memory_shares ADD COLUMN IF NOT EXISTS target_memory_id UUID REFERENCES team_memories(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_memory_shares_target_memory ON memory_shares(target_memory_id) WHERE target_memory_id IS NOT NULL;

COMMENT ON COLUMN memory_shares.target_memory_id IS 'Reference to the memory record created in the target team (for revocation and sync)';

-- Update the memory sharing manager to populate this field when accepting shares
-- Existing shares will need to be backfilled or left NULL (revoke will use search fallback)
