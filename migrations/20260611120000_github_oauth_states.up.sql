-- Add user_id and tenant_id to oauth_states for GitHub OAuth support
ALTER TABLE oauth_states ADD COLUMN IF NOT EXISTS user_id UUID;
ALTER TABLE oauth_states ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE oauth_states ADD COLUMN IF NOT EXISTS provider VARCHAR(50) DEFAULT 'github';

-- Existing rows won't have user_id/tenant_id, so allow nulls initially
-- But we want them NOT NULL for new entries, so add check constraint
ALTER TABLE oauth_states ADD CONSTRAINT chk_oauth_states_user_id CHECK (provider != 'github' OR user_id IS NOT NULL);
ALTER TABLE oauth_states ADD CONSTRAINT chk_oauth_states_tenant_id CHECK (provider != 'github' OR tenant_id IS NOT NULL);

-- Index for querying by user_id (for cleanup/validation)
CREATE INDEX IF NOT EXISTS idx_oauth_states_user_id ON oauth_states(user_id) WHERE user_id IS NOT NULL;