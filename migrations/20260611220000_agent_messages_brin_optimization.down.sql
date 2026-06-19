-- Revert agent_messages BRIN optimization
-- Migration: 20260611220000_agent_messages_brin_optimization

-- Drop BRIN indexes
DROP INDEX IF EXISTS idx_agent_messages_created_at_brin;
DROP INDEX IF EXISTS idx_agent_messages_to_agent_brin;
DROP INDEX IF EXISTS idx_agent_messages_pending_brin;

-- Recreate original B-tree indexes
CREATE INDEX IF NOT EXISTS idx_agent_messages_from_agent ON agent_messages (from_agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_messages_to_agent ON agent_messages (to_agent_id, status, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_agent_messages_pending_inbox ON agent_messages (to_agent_id, created_at ASC)
    WHERE status IN ('pending', 'delivered');
CREATE INDEX IF NOT EXISTS idx_agent_messages_nonce ON agent_messages (nonce);