-- Revert: Recreate redundant indexes dropped by 20260616220001_consolidate_agent_messages_indexes
--
-- WARNING: This reverts to the state with redundant indexes, which causes slow INSERTs.
-- Only use for testing/rollback, not for production.

-- Recreate BRIN indexes
CREATE INDEX IF NOT EXISTS idx_agent_messages_created_at_brin ON agent_messages USING brin (created_at);
CREATE INDEX IF NOT EXISTS idx_agent_messages_to_agent_brin ON agent_messages USING brin (to_agent_id, created_at)
    WITH (pages_per_range = 4);
CREATE INDEX IF NOT EXISTS idx_agent_messages_pending_brin ON agent_messages USING brin (to_agent_id, created_at)
    WHERE status IN ('pending', 'delivered');

-- Recreate B-tree partial indexes
CREATE INDEX IF NOT EXISTS idx_agent_messages_to_agent_pending
    ON agent_messages(to_agent_id, created_at ASC)
    WHERE delivered_at IS NULL AND read_at IS NULL;

-- Recreate idx_agent_messages_from_agent (note: this was created before outbox_btree)
CREATE INDEX IF NOT EXISTS idx_agent_messages_from_agent
    ON agent_messages(from_agent_id, created_at DESC);
