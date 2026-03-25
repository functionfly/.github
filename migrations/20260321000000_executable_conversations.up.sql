-- Executable Conversations: DM/conversation layer with function-aware messages
-- Conversation types: dm, function_thread, issue_thread, fix_mode, bounty_thread, org_thread, security_disclosure
-- Idempotent: type may already exist if migration was re-run after state repair.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'conversation_type_enum') THEN
    CREATE TYPE conversation_type_enum AS ENUM (
      'dm',
      'function_thread',
      'issue_thread',
      'fix_mode',
      'bounty_thread',
      'org_thread',
      'security_disclosure'
    );
  END IF;
END
$$;

CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type conversation_type_enum NOT NULL DEFAULT 'dm',

    -- Participants: array of user UUIDs as JSONB array of strings (for DMs typically 2; for org_thread can be more)
    participant_ids JSONB NOT NULL DEFAULT '[]'::jsonb,

    -- Optional link to Flywheel thread (e.g. when "Move to Private Debug Thread")
    source_thread_id UUID REFERENCES flywheel_threads(id),
    -- Optional org/tenant scope for org_thread
    organization_id UUID,

    metadata JSONB DEFAULT '{}'::jsonb,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_participant_ids ON conversations USING GIN(participant_ids);
COMMENT ON COLUMN conversations.participant_ids IS 'JSON array of UUID strings, e.g. ["uuid1","uuid2"]';
CREATE INDEX IF NOT EXISTS idx_conversations_type ON conversations(type);
CREATE INDEX IF NOT EXISTS idx_conversations_source_thread ON conversations(source_thread_id);
CREATE INDEX IF NOT EXISTS idx_conversations_updated_at ON conversations(updated_at DESC);

-- Messages with embeddings for function ref, execution snippet, etc.
CREATE TABLE IF NOT EXISTS conversation_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id),

    content TEXT NOT NULL DEFAULT '',
    -- Embeddings: { "function_ref"?: { "author", "name", "version" }, "execution_id"?: "uuid", "execution_root_hash"?: "0x...", "replay_link"?: "url", "capability_declaration"?: {}, "input_summary"?, "output_summary"? }
    embeddings JSONB DEFAULT '{}'::jsonb,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversation_messages_conversation_id ON conversation_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversation_messages_author_id ON conversation_messages(author_id);
CREATE INDEX IF NOT EXISTS idx_conversation_messages_created_at ON conversation_messages(conversation_id, created_at DESC);
