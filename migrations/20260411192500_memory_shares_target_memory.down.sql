-- Migration: Remove target_memory_id column from memory_shares
-- Reverses: 20260411192500_memory_shares_target_memory

-- Drop the index first
DROP INDEX IF EXISTS idx_memory_shares_target_memory;

-- Remove the column
ALTER TABLE memory_shares DROP COLUMN IF EXISTS target_memory_id;