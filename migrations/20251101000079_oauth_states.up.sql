-- OAuth state table for CSRF protection (persisted for multi-instance OAuth flows)
CREATE TABLE IF NOT EXISTS oauth_states (
    state VARCHAR(512) PRIMARY KEY,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_oauth_states_expires_at ON oauth_states(expires_at);
