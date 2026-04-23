-- Rollback conversation enhancements

-- 4. Drop attachments table
DROP TABLE IF EXISTS message_attachments;

-- 3. Remove read cursor columns
DROP INDEX IF EXISTS idx_conversation_participant_reads_last_msg;
ALTER TABLE conversation_participant_reads DROP COLUMN IF EXISTS last_read_message_id;

-- 2. Remove full-text search
DROP INDEX IF EXISTS idx_conversation_messages_content_search;
ALTER TABLE conversation_messages DROP COLUMN IF EXISTS content_search;

-- 1. Remove edit/delete columns
DROP INDEX IF EXISTS idx_conversation_messages_deleted_at;
ALTER TABLE conversation_messages DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE conversation_messages DROP COLUMN IF EXISTS edited_at;
