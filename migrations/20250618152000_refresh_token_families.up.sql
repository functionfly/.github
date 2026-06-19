-- Add family_id for refresh token rotation with device families
-- Each family represents a unique device/browser combination

ALTER TABLE refresh_tokens ADD COLUMN IF NOT EXISTS family_id UUID DEFAULT gen_random_uuid();

-- Index for looking up tokens by family (for session management)
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family ON refresh_tokens(family_id) WHERE family_id IS NOT NULL;

-- Index for finding active token in a family (most recent non-revoked)
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family_active ON refresh_tokens(family_id, created_at DESC) WHERE revoked = false;

COMMENT ON COLUMN refresh_tokens.family_id IS 'Device family identifier. All refresh tokens from the same device share a family_id. Used for rotation replay attack detection.';
