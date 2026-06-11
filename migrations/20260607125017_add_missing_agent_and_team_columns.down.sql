-- Rollback: remove columns from agent_messages
ALTER TABLE agent_messages DROP COLUMN IF EXISTS signature;
ALTER TABLE agent_messages DROP COLUMN IF EXISTS nonce;
ALTER TABLE agent_messages DROP COLUMN IF EXISTS sequence_number;
ALTER TABLE agent_messages DROP COLUMN IF EXISTS trace_id;
ALTER TABLE agent_messages DROP COLUMN IF EXISTS span_id;
ALTER TABLE agent_messages DROP COLUMN IF EXISTS trace_flags;
ALTER TABLE agent_messages DROP COLUMN IF EXISTS trace_state;

-- Rollback: remove columns from teams
ALTER TABLE teams DROP COLUMN IF EXISTS slug;
ALTER TABLE teams DROP COLUMN IF EXISTS visibility;
