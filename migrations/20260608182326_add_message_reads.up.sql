-- Migration: Add message read receipts table
-- Tracks individual message reads (who read which message and when)

CREATE TABLE IF NOT EXISTS conversation_message_reads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES conversation_messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    read_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(message_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_message_reads_message_id ON conversation_message_reads(message_id);
CREATE INDEX IF NOT EXISTS idx_message_reads_user_id ON conversation_message_reads(user_id);
CREATE INDEX IF NOT EXISTS idx_message_reads_read_at ON conversation_message_reads(read_at);

-- Down migration
DROP TABLE IF EXISTS conversation_message_reads;