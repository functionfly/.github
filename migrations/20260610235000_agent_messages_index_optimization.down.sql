-- Revert agent_messages index optimization
-- Migration: 20260610235000_agent_messages_index_optimization

-- Drop the new optimized indexes
DROP INDEX IF EXISTS idx_agent_messages_id_status;
DROP INDEX IF EXISTS idx_agent_messages_pending_inbox;
DROP INDEX IF EXISTS idx_agent_messages_to_agent;

-- Recreate original indexes
CREATE INDEX IF NOT EXISTS idx_agent_messages_to_agent ON agent_messages(to_agent_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_messages_session ON agent_messages(session_id);
