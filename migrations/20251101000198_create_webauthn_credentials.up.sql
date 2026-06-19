-- Create webauthn_credentials table for WebAuthn/Passkeys authentication
CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id BYTEA NOT NULL,
    public_key BYTEA NOT NULL,
    sign_count INTEGER NOT NULL DEFAULT 0,
    backup_eligible BOOLEAN DEFAULT false,
    backup_state BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP
);

-- Index for faster lookups by user_id
CREATE INDEX idx_webauthn_user_id ON webauthn_credentials(user_id);

-- Index for credential_id lookups (needed for authentication)
CREATE INDEX idx_webauthn_credential_id ON webauthn_credentials(credential_id);
