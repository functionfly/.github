-- +up
-- Add message reactions support for emoji reactions on conversation messages.

CREATE TABLE message_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES conversation_messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction VARCHAR(50) NOT NULL, -- emoji Unicode or shortcode like :smile:
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(message_id, user_id, reaction)
);

CREATE INDEX idx_message_reactions_message_id ON message_reactions(message_id);
CREATE INDEX idx_message_reactions_user_id ON message_reactions(user_id);

-- +down
DROP TABLE IF EXISTS message_reactions;
