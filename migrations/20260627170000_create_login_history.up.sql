-- Create login_history table for tracking user sign-in events
CREATE TABLE IF NOT EXISTS login_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL DEFAULT 'login', -- 'login', 'logout', 'logout_other', 'session_expired', 'revoked'
    ip_address INET NOT NULL,
    user_agent TEXT,
    device VARCHAR(255),
    location VARCHAR(255), -- Optional: derived from IP
    login_method VARCHAR(50), -- 'password', 'google', 'github', 'saml', 'webauthn', etc.
    mfa_used BOOLEAN NOT NULL DEFAULT FALSE,
    session_id UUID, -- Reference to the session if still active
    metadata JSONB, -- Additional context (failure reason, etc.)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_login_history_user_id ON login_history(user_id);
CREATE INDEX IF NOT EXISTS idx_login_history_created_at ON login_history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_history_user_created ON login_history(user_id, created_at DESC);

-- Retention policy: keep login history for 90 days (configurable)
-- This aligns with the detailed execution logs retention policy
