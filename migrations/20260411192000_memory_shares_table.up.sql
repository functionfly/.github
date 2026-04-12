-- Migration: Create memory_shares table for cross-team memory sharing
-- Created: 2026-04-11

CREATE TABLE IF NOT EXISTS memory_shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    memory_id UUID NOT NULL REFERENCES team_memories(id) ON DELETE CASCADE,
    source_team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    target_team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    shared_by UUID NOT NULL REFERENCES users(id),
    
    share_type VARCHAR(20) NOT NULL DEFAULT 'reference', -- 'reference', 'copy', 'fork'
    permission VARCHAR(20) NOT NULL DEFAULT 'read',     -- 'read', 'write', 'admin'
    status VARCHAR(20) NOT NULL DEFAULT 'pending',     -- 'pending', 'accepted', 'rejected', 'revoked'
    
    message TEXT,
    accepted_by UUID REFERENCES users(id),
    accepted_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Prevent duplicate shares
    CONSTRAINT unique_memory_team_share UNIQUE (memory_id, source_team_id, target_team_id),
    -- Prevent self-sharing
    CONSTRAINT no_self_share CHECK (source_team_id != target_team_id)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_memory_shares_memory_id ON memory_shares(memory_id);
CREATE INDEX IF NOT EXISTS idx_memory_shares_source_team ON memory_shares(source_team_id);
CREATE INDEX IF NOT EXISTS idx_memory_shares_target_team ON memory_shares(target_team_id);
CREATE INDEX IF NOT EXISTS idx_memory_shares_status ON memory_shares(status);
CREATE INDEX IF NOT EXISTS idx_memory_shares_pending_target ON memory_shares(target_team_id, status) WHERE status = 'pending';

-- RLS policies for memory_shares
ALTER TABLE memory_shares ENABLE ROW LEVEL SECURITY;

-- Team members can see shares involving their teams
CREATE POLICY memory_shares_team_access ON memory_shares
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM team_memberships tm
            WHERE tm.user_id = current_user_id()
            AND (tm.team_id = memory_shares.source_team_id OR tm.team_id = memory_shares.target_team_id)
        )
    );

-- Auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION update_memory_shares_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS memory_shares_updated_at_trigger ON memory_shares;
CREATE TRIGGER memory_shares_updated_at_trigger
    BEFORE UPDATE ON memory_shares
    FOR EACH ROW EXECUTE FUNCTION update_memory_shares_updated_at();

-- Comments
COMMENT ON TABLE memory_shares IS 'Cross-team memory sharing - enables collaboration and knowledge transfer between teams';
COMMENT ON COLUMN memory_shares.share_type IS 'reference = synced copy, copy = independent, fork = tracked copy';
COMMENT ON COLUMN memory_shares.permission IS 'read = view only, write = can suggest updates, admin = full control';
