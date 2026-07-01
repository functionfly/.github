-- Migration: add_agent_id_to_chat_sessions
-- Purpose: Add agent_id column to ai_chat_sessions for per-agent conversation history.

BEGIN;

ALTER TABLE ai_chat_sessions ADD COLUMN IF NOT EXISTS agent_id TEXT;
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_agent ON ai_chat_sessions(agent_id);

COMMIT;
