-- Migration: Add index on state_fabric_snapshots.expires_at for TTL cleanup worker
-- Created: 2026-04-19

-- Add index for efficient TTL cleanup queries on expired snapshots
CREATE INDEX IF NOT EXISTS idx_state_fabric_snapshots_expires_at 
    ON state_fabric_snapshots(expires_at) 
    WHERE expires_at IS NOT NULL;
