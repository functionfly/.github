-- Migration: Drop memory_shares table
-- Reverses: 20260411192000_memory_shares_table

-- Drop trigger first
DROP TRIGGER IF EXISTS memory_shares_updated_at_trigger ON memory_shares;

-- Drop helper function
DROP FUNCTION IF EXISTS update_memory_shares_updated_at();

-- Drop table
DROP TABLE IF EXISTS memory_shares;