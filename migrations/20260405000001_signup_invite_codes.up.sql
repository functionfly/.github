-- Create signup_invite_codes table for invite-only beta launch.
CREATE TABLE IF NOT EXISTS signup_invite_codes (
    id UUID PRIMARY KEY,
    code_fingerprint VARCHAR(64) NOT NULL UNIQUE,
    code_hash TEXT NOT NULL,
    label VARCHAR(512),
    max_uses INT,
    uses_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_signup_invite_fingerprint ON signup_invite_codes (code_fingerprint);
CREATE INDEX IF NOT EXISTS idx_signup_invite_created_by ON signup_invite_codes (created_by);

-- Add invite_code column to oauth_states for carrying invite codes through the OAuth flow.
ALTER TABLE oauth_states ADD COLUMN IF NOT EXISTS invite_code TEXT;
