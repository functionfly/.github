-- =====================================================
-- Add Trust API fields to the unified api_keys table
-- =====================================================

-- Add partner_id column for Trust API keys (references trust_api_partners)
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS partner_id UUID;

-- Add scopes for Trust API keys (JSONB map)
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS scopes JSONB DEFAULT '{"trust:read": true}'::jsonb;

-- Add revocation fields
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS is_revoked BOOLEAN DEFAULT FALSE;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS revoked_reason TEXT;

-- Add use_count for Trust API keys
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS use_count INTEGER DEFAULT 0;

-- Add allowed_ips for Trust API keys
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS allowed_ips JSONB DEFAULT '[]'::jsonb;

-- Add created_by for Trust API keys
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) DEFAULT '';

-- Add key_id column (public key identifier) if not exists
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_id VARCHAR(64) UNIQUE;

-- Add indexes for the new columns
CREATE INDEX IF NOT EXISTS idx_api_keys_revoked ON api_keys(is_revoked) WHERE is_revoked = TRUE;
CREATE INDEX IF NOT EXISTS idx_api_keys_partner ON api_keys(partner_id) WHERE partner_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_key_id ON api_keys(key_id) WHERE key_id IS NOT NULL;

-- Update the CHECK constraint to allow 'trust' key type
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_key_type_check;
ALTER TABLE api_keys ADD CONSTRAINT api_keys_key_type_check
    CHECK (key_type IN ('platform', 'function', 'agent', 'environment', 'oauth', 'trust'));

-- Update key_prefix comment to include fft_
COMMENT ON COLUMN api_keys.key_prefix IS 'Key type prefix (ffp_, fff_, aep_, ffe_, ffo_, fft_)';
COMMENT ON COLUMN api_keys.key_type IS 'Type of API key: platform, function, agent, environment, oauth, or trust';
COMMENT ON COLUMN api_keys.partner_id IS 'Trust partner ID for Trust API keys';
COMMENT ON COLUMN api_keys.scopes IS 'Scopes for Trust API keys (JSONB map)';
COMMENT ON COLUMN api_keys.is_revoked IS 'Whether the Trust API key has been revoked';
COMMENT ON COLUMN api_keys.revoked_at IS 'Timestamp when the key was revoked';
COMMENT ON COLUMN api_keys.revoked_reason IS 'Reason for revocation';
COMMENT ON COLUMN api_keys.use_count IS 'Request count for Trust API keys';
COMMENT ON COLUMN api_keys.allowed_ips IS 'Allowed IP addresses for Trust API keys (JSONB array)';
COMMENT ON COLUMN api_keys.created_by IS 'User or system that created the key';
COMMENT ON COLUMN api_keys.key_id IS 'Public key identifier (full key for Trust, prefix+hash for others)';
