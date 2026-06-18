-- Migration: 20260616164702_add_btree_index_agent_messages_inbox
--
-- Problem: BRIN indexes on (to_agent_id, created_at) are ineffective because
--          to_agent_id is high-cardinality and randomly distributed across pages.
--          BRIN min/max ranges become too wide, causing poor query performance.
--
-- Solution: Add a B-tree composite index that supports the inbox query:
--   WHERE to_agent_id = ? AND status IN ('pending', 'delivered')
--   ORDER BY created_at ASC
--
-- The composite index (to_agent_id, status, created_at) allows:
--   1. Direct equality lookup on to_agent_id
--   2. Filtering by status via index scan
--   3. Ordered retrieval by created_at without sorting

-- Add B-tree index for inbox queries (to_agent_id filter + status filter + created_at ordering)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_agent_messages_inbox_btree
    ON agent_messages (to_agent_id, status, created_at ASC)
    WHERE status IN ('pending', 'delivered');

-- Also add a B-tree index for outbox queries (from_agent_id + created_at ordering)
-- This supports: WHERE from_agent_id = ? ORDER BY created_at DESC
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_agent_messages_outbox_btree
    ON agent_messages (from_agent_id, created_at DESC);
