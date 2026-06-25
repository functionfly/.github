-- Rollback: 20260619182100_fix_agent_messages_invalid_indexes
--
-- This migration cannot undo the index recreation, but we mark it as applied
-- The actual fix is idempotent - running it again will verify and fix if needed

-- No-op: we don't drop the indexes since they are needed for performance
-- The ANALYZE is also not reversible, but it's a low-impact operation
