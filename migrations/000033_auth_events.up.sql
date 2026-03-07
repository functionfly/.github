-- Authentication events table for security auditing and compliance
CREATE TABLE IF NOT EXISTS auth_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,

    event_type VARCHAR(50) NOT NULL, -- login, logout, mfa_verify, oauth_login, password_reset, etc.
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100), -- invalid_credentials, account_locked, mfa_failed, etc.

    ip_address INET,
    user_agent TEXT,
    location_info JSONB, -- geolocation data if available

    session_id UUID, -- for session tracking
    provider VARCHAR(50), -- oauth provider (github, google, etc.)

    metadata JSONB, -- additional context (device info, risk scores, etc.)
    security_flags JSONB, -- security-related flags (suspicious_login, new_device, etc.)

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance and querying
CREATE INDEX IF NOT EXISTS idx_auth_events_user_id ON auth_events(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_events_tenant_id ON auth_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_auth_events_event_type ON auth_events(event_type);
CREATE INDEX IF NOT EXISTS idx_auth_events_success ON auth_events(success);
CREATE INDEX IF NOT EXISTS idx_auth_events_created_at ON auth_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_auth_events_ip_address ON auth_events(ip_address);
CREATE INDEX IF NOT EXISTS idx_auth_events_session_id ON auth_events(session_id);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_auth_events_user_type_created ON auth_events(user_id, event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_auth_events_ip_recent ON auth_events(ip_address, created_at DESC);

-- Partial indexes for failed events
CREATE INDEX IF NOT EXISTS idx_auth_events_failed_recent ON auth_events(created_at DESC) WHERE success = false;
CREATE INDEX IF NOT EXISTS idx_auth_events_failed_user ON auth_events(user_id, created_at DESC) WHERE success = false;