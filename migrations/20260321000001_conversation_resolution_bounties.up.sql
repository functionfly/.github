-- Resolution and bounties for executable conversations

-- Resolution fields on conversations
ALTER TABLE conversations
  ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS resolved_by_user_id UUID REFERENCES users(id),
  ADD COLUMN IF NOT EXISTS resolved_by_message_id UUID REFERENCES conversation_messages(id);

CREATE INDEX IF NOT EXISTS idx_conversations_resolved_at ON conversations(resolved_at) WHERE resolved_at IS NOT NULL;

-- Bounties attached to conversations
CREATE TABLE IF NOT EXISTS conversation_bounties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    offered_by UUID NOT NULL REFERENCES users(id),

    amount_reputation INT NOT NULL DEFAULT 0 CHECK (amount_reputation >= 0),
    amount_cents INT DEFAULT 0 CHECK (amount_cents >= 0),
    security_weight_multiplier NUMERIC(5, 2) DEFAULT 1.0,

    claimed_by UUID REFERENCES users(id),
    claimed_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversation_bounties_conversation ON conversation_bounties(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conversation_bounties_offered_by ON conversation_bounties(offered_by);
CREATE INDEX IF NOT EXISTS idx_conversation_bounties_claimed ON conversation_bounties(claimed_by) WHERE claimed_by IS NOT NULL;
