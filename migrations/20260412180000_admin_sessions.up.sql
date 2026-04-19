-- Migration: Create admin_sessions table for session tracking
-- This table stores server-side session validation data for admin users

CREATE TABLE IF NOT EXISTS admin_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Session identification
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    device_fingerprint VARCHAR(128),
    user_agent TEXT,

    -- Timing
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    last_activity_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,

    -- Security
    ip_address INET NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE NOT NULL,
    revoked_at TIMESTAMPTZ,

    -- Metadata
    fingerprint_mismatch_warnings INTEGER DEFAULT 0
);

-- Indexes for efficient session lookups
CREATE INDEX idx_admin_sessions_user_id ON admin_sessions(user_id);
CREATE INDEX idx_admin_sessions_token_hash ON admin_sessions(token_hash);
CREATE INDEX idx_admin_sessions_expires_at ON admin_sessions(expires_at);
CREATE INDEX idx_admin_sessions_last_activity ON admin_sessions(last_activity_at);

-- Comment on table
COMMENT ON TABLE admin_sessions IS 'Server-side admin session tracking for enhanced security';
COMMENT ON COLUMN admin_sessions.token_hash IS 'SHA256 hash of the session token (never store plaintext)';
COMMENT ON COLUMN admin_sessions.device_fingerprint IS 'Hashed device fingerprint for device validation';
COMMENT ON COLUMN admin_sessions.ip_address IS 'Client IP address at session creation';
COMMENT ON COLUMN admin_sessions.fingerprint_mismatch_warnings IS 'Count of fingerprint mismatches (new device indicator)';
