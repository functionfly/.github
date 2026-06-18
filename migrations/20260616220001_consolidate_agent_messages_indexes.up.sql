-- Migration: 20260616220001_consolidate_agent_messages_indexes
--
-- Problem: Multiple overlapping indexes on agent_messages causing 200+ms INSERT latency
--          Indexes were added in multiple passes without removing redundant ones.
--
-- Root cause analysis:
--   - 20260611220000 added BRIN indexes (replaced B-tree for INSERT performance)
--   - 20260615000003 added partial B-tree indexes
--   - 20260616164702 added B-tree inbox/outbox indexes
--   Result: 8+ indexes on same columns, all updated on every INSERT
--
-- Solution: Keep minimal set of B-tree indexes, drop redundant BRIN and partial indexes
--
-- Keep (3 indexes):
--   1. agent_messages_pkey - primary key (required)
--   2. idx_agent_messages_inbox_btree - (to_agent_id, status, created_at ASC)
--      WHERE status IN ('pending', 'delivered') - for inbox polling
--   3. idx_agent_messages_outbox_btree - (from_agent_id, created_at DESC) - for outbox queries
--
-- Drop (5 redundant indexes):
--   1. idx_agent_messages_created_at_brin - BRIN on created_at (overlaps outbox_btree)
--   2. idx_agent_messages_to_agent_brin - BRIN on (to_agent_id, created_at) (overlaps inbox_btree)
--   3. idx_agent_messages_pending_brin - BRIN partial (overlaps inbox_btree)
--   4. idx_agent_messages_to_agent_pending - B-tree partial (overlaps inbox_btree)
--   5. idx_agent_messages_from_agent - B-tree (overlaps outbox_btree)
--
-- Expected improvement: ~200ms → ~5ms per INSERT (40x improvement)

-- Step 1: Drop redundant BRIN indexes
DROP INDEX IF EXISTS idx_agent_messages_created_at_brin;
DROP INDEX IF EXISTS idx_agent_messages_to_agent_brin;
DROP INDEX IF EXISTS idx_agent_messages_pending_brin;

-- Step 2: Drop redundant B-tree indexes (partial and full)
DROP INDEX IF EXISTS idx_agent_messages_to_agent_pending;
DROP INDEX IF EXISTS idx_agent_messages_from_agent;

-- Note: idx_agent_messages_to_agent_read is not dropped because it may not exist
-- (CREATE INDEX IF NOT EXISTS would have failed if read_at column was missing)
DROP INDEX IF EXISTS idx_agent_messages_to_agent_read;

-- Retained indexes (created by 20260616164702):
--   - idx_agent_messages_inbox_btree (to_agent_id, status, created_at ASC) WHERE status IN ('pending', 'delivered')
--   - idx_agent_messages_outbox_btree (from_agent_id, created_at DESC)
--   - agent_messages_pkey (id)
