-- Per-user read cursor for executable conversations (sidebar unread badges).

CREATE TABLE IF NOT EXISTS conversation_participant_reads (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_conversation_participant_reads_user
    ON conversation_participant_reads (user_id);

COMMENT ON TABLE conversation_participant_reads IS 'Last time each participant viewed a conversation; used for unread message counts.';
