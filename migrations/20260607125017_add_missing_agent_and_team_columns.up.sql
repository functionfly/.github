-- Add missing columns to teams table
ALTER TABLE teams ADD COLUMN IF NOT EXISTS slug VARCHAR(100);
ALTER TABLE teams ADD COLUMN IF NOT EXISTS visibility VARCHAR(20) DEFAULT 'private';

-- Add missing columns to agent_messages table
ALTER TABLE agent_messages ADD COLUMN IF NOT EXISTS signature TEXT;
ALTER TABLE agent_messages ADD COLUMN IF NOT EXISTS nonce TEXT;
ALTER TABLE agent_messages ADD COLUMN IF NOT EXISTS sequence_number BIGINT DEFAULT 0;
ALTER TABLE agent_messages ADD COLUMN IF NOT EXISTS trace_id TEXT;
ALTER TABLE agent_messages ADD COLUMN IF NOT EXISTS span_id TEXT;
ALTER TABLE agent_messages ADD COLUMN IF NOT EXISTS trace_flags TEXT;
ALTER TABLE agent_messages ADD COLUMN IF NOT EXISTS trace_state TEXT;

-- Add index on nonce for agent_messages
CREATE INDEX IF NOT EXISTS idx_agent_messages_nonce ON agent_messages(nonce);
