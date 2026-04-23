-- Conversation enhancements: message edit/delete, full-text search, read cursor fix, attachments

-- 1. Message editing and deletion
ALTER TABLE conversation_messages ADD COLUMN IF NOT EXISTS edited_at TIMESTAMPTZ;
ALTER TABLE conversation_messages ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_conversation_messages_deleted_at ON conversation_messages(deleted_at) WHERE deleted_at IS NOT NULL;

-- 2. Full-text search on message content
ALTER TABLE conversation_messages ADD COLUMN IF NOT EXISTS content_search TSVECTOR
    GENERATED ALWAYS AS (to_tsvector('english', content)) STORED;
CREATE INDEX IF NOT EXISTS idx_conversation_messages_content_search ON conversation_messages USING GIN(content_search);

-- 3. Read cursor granularity: add last_read_message_id alongside last_read_at
ALTER TABLE conversation_participant_reads ADD COLUMN IF NOT EXISTS last_read_message_id UUID REFERENCES conversation_messages(id);
CREATE INDEX IF NOT EXISTS idx_conversation_participant_reads_last_msg ON conversation_participant_reads(last_read_message_id);

-- 4. Message attachments table
CREATE TABLE IF NOT EXISTS message_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES conversation_messages(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    size_bytes BIGINT NOT NULL DEFAULT 0,
    storage_url TEXT NOT NULL,
    thumbnail_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_message_attachments_message_id ON message_attachments(message_id);
CREATE INDEX IF NOT EXISTS idx_message_attachments_conversation_id ON message_attachments(conversation_id);
