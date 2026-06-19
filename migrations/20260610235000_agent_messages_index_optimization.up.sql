-- Optimize agent_messages indexes for query performance
-- Migration: 20260610235000_agent_messages_index_optimization
--
-- Issue: Slow INSERTs (200-350ms) due to multiple indexes being updated
-- Solution:
--   1. Fix idx_agent_messages_to_agent to use ASC to match query ORDER BY
--   2. Drop rarely-used idx_agent_messages_session index

-- Step 1: Drop the incorrectly ordered to_agent_id index
DROP INDEX IF EXISTS idx_agent_messages_to_agent;

-- Step 2: Drop the rarely-used session index (can be served by other indexes if needed)
DROP INDEX IF EXISTS idx_agent_messages_session;

-- Step 3: Recreate idx_agent_messages_to_agent with correct ASC order
-- This matches the query: ORDER BY created_at ASC
CREATE INDEX IF NOT EXISTS idx_agent_messages_to_agent ON agent_messages(to_agent_id, status, created_at ASC);

-- Step 4: Create a partial index for pending messages only (smaller, faster for common queries)
-- This covers the most common inbox query pattern with less overhead
CREATE INDEX IF NOT EXISTS idx_agent_messages_pending_inbox ON agent_messages(to_agent_id, created_at ASC)
    WHERE status IN ('pending', 'delivered');

-- Step 5: Create a covering index for MarkRead/MarkDelivered (id lookups)
-- The id is already the primary key, but this helps with the status check
CREATE INDEX IF NOT EXISTS idx_agent_messages_id_status ON agent_messages(id, status)
    WHERE status = 'pending';
