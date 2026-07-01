BEGIN;

DROP INDEX IF EXISTS idx_ai_chat_sessions_agent;
ALTER TABLE ai_chat_sessions DROP COLUMN IF EXISTS agent_id;

COMMIT;
