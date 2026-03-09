-- GoBetterAuth WebAuthn Tables
-- Creates tables for WebAuthn/Passkey authentication

-- WebAuthn credentials table for storing passkeys
CREATE TABLE IF NOT EXISTS gba_webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES gba_users(id) ON DELETE CASCADE,
    credential_id BYTEA NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    attestation_type VARCHAR(50) NOT NULL,
    transport TEXT[], -- Array of transport methods
    flags INT NOT NULL DEFAULT 0,
    authenticator JSONB NOT NULL,
    sign_count INT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT 'Passkey',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- WebAuthn sessions table for storing registration/authentication sessions
CREATE TABLE IF NOT EXISTS gba_webauthn_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES gba_users(id) ON DELETE CASCADE,
    challenge VARCHAR(255) NOT NULL,
    operation VARCHAR(20) NOT NULL CHECK (operation IN ('registration', 'authentication')),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_gba_webauthn_creds_user ON gba_webauthn_credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_gba_webauthn_creds_cred_id ON gba_webauthn_credentials(credential_id);
CREATE INDEX IF NOT EXISTS idx_gba_webauthn_sessions_user ON gba_webauthn_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_gba_webauthn_sessions_expires ON gba_webauthn_sessions(expires_at);

-- Create a partial index for active credentials
CREATE INDEX IF NOT EXISTS idx_gba_webauthn_creds_active 
ON gba_webauthn_credentials(user_id) 
WHERE deleted_at IS NULL;

-- Add comments for documentation
COMMENT ON TABLE gba_webauthn_credentials IS 'WebAuthn/Passkey credentials for GoBetterAuth users';
COMMENT ON TABLE gba_webauthn_sessions IS 'Temporary sessions for WebAuthn registration and authentication ceremonies';
COMMENT ON COLUMN gba_webauthn_credentials.credential_id IS 'Unique identifier for the WebAuthn credential';
COMMENT ON COLUMN gba_webauthn_credentials.public_key IS 'Public key associated with the credential';
COMMENT ON COLUMN gba_webauthn_credentials.sign_count IS 'Signature counter to prevent replay attacks';
COMMENT ON COLUMN gba_webauthn_sessions.challenge IS 'Base64-encoded challenge used in the WebAuthn ceremony';
COMMENT ON COLUMN gba_webauthn_sessions.operation IS 'Type of WebAuthn operation: registration or authentication';