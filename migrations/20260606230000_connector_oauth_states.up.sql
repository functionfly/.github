-- Create connector_oauth_states table for CSRF protection
CREATE TABLE IF NOT EXISTS connector_oauth_states (
    state VARCHAR(64) PRIMARY KEY,
    tenant_id UUID NOT NULL,
    connector_id UUID NOT NULL REFERENCES connectors(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '10 minutes',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for efficient cleanup of expired states
CREATE INDEX IF NOT EXISTS idx_connector_oauth_states_expires ON connector_oauth_states(expires_at);

-- Index for tenant lookups
CREATE INDEX IF NOT EXISTS idx_connector_oauth_states_tenant ON connector_oauth_states(tenant_id);