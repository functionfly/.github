-- Rollback: add_ai_chat_tables
-- Created at: 2026-06-30T17:51:02-05:00

BEGIN;

DROP TABLE IF EXISTS ai_chat_messages;
DROP TABLE IF EXISTS ai_chat_sessions;

COMMIT;
