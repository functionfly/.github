-- Migration: Add trusted device token expiry to sessions table
-- Previously trusted device tokens had no expiry - they could be reused indefinitely.
-- Now we track when the trusted device relationship itself expires (typically 30 days).

-- Add the trusted device token expiry column (may already exist from partial run)
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS trusted_device_token_expires_at TIMESTAMPTZ;

-- Note: trusted_device_token column does not exist in this schema.
-- The trusted device feature relies on session token binding in application code.
-- trusted_device_token_expires_at is set when creating trusted sessions.

-- Add index for efficient lookup of trusted token expiry queries
CREATE INDEX IF NOT EXISTS idx_sessions_trusted_token_expires
ON sessions(trusted_device_token_expires_at)
WHERE trusted_device_token_expires_at IS NOT NULL;

COMMENT ON COLUMN sessions.trusted_device_token_expires_at IS 'Expiry for the trusted device token itself. If NULL, token never expires. If set, must be checked before accepting the token.';