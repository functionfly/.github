-- GoBetterAuth Base Tables
-- Creates the core tables for GoBetterAuth authentication system

-- Tenants table for multi-tenancy
CREATE TABLE IF NOT EXISTS gba_tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(255) UNIQUE,
    status VARCHAR(50) DEFAULT 'active',
    mfa_policy VARCHAR(50) DEFAULT 'optional',
    session_max_duration INTEGER DEFAULT 604800,
    session_idle_timeout INTEGER DEFAULT 1800,
    concurrent_sessions_limit INTEGER DEFAULT 5,
    allowed_email_domains TEXT[],
    oauth_providers_enabled TEXT[],
    ip_allowlist_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Users table for authentication
CREATE TABLE IF NOT EXISTS gba_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255),
    name VARCHAR(255),
    image VARCHAR(512),
    role VARCHAR(50) DEFAULT 'user',
    email_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP WITH TIME ZONE,
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255),
    provider VARCHAR(50),
    provider_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Accounts table for OAuth providers
CREATE TABLE IF NOT EXISTS gba_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES gba_users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_account_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    token_type VARCHAR(50),
    scope VARCHAR(512),
    expires_at TIMESTAMP WITH TIME ZONE,
    session_state VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Sessions table for user sessions
CREATE TABLE IF NOT EXISTS gba_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES gba_users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    mfa_verified BOOLEAN DEFAULT FALSE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_gba_users_email_tenant ON gba_users(email, tenant_id);
CREATE INDEX IF NOT EXISTS idx_gba_users_tenant ON gba_users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gba_accounts_user ON gba_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_gba_accounts_provider ON gba_accounts(provider, provider_account_id);
CREATE INDEX IF NOT EXISTS idx_gba_sessions_user ON gba_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_gba_sessions_token ON gba_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_gba_sessions_expires ON gba_sessions(expires_at);

-- Add comments for documentation
COMMENT ON TABLE gba_tenants IS 'Multi-tenant organizations for GoBetterAuth';
COMMENT ON TABLE gba_users IS 'User accounts for authentication';
COMMENT ON TABLE gba_accounts IS 'OAuth provider accounts linked to users';
COMMENT ON TABLE gba_sessions IS 'Active user authentication sessions';
