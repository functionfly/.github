-- Migration: Remove index on state_fabric_snapshots.expires_at
-- Created: 2026-04-19

DROP INDEX IF EXISTS idx_state_fabric_snapshots_expires_at;
