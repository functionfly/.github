-- Create tenant_auth_settings table for per-tenant auth configuration
-- This supports Backend-in-a-Box bundles with configurable auth per tenant

CREATE TABLE IF NOT EXISTS tenant_auth_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID UNIQUE NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    mfa_required BOOLEAN NOT NULL DEFAULT false,
    mfa_mode VARCHAR(20) NOT NULL DEFAULT 'optional' CHECK (mfa_mode IN ('optional', 'required', 'enforced')),
    password_policy JSONB NOT NULL DEFAULT '{"min_length": 8, "require_uppercase": true, "require_lowercase": true, "require_digit": true, "require_special": true}',
    session_timeout_minutes INT NOT NULL DEFAULT 480,
    ip_allowlist_enabled BOOLEAN NOT NULL DEFAULT false,
    ip_allowlist JSONB NOT NULL DEFAULT '[]',
    allowed_domains JSONB NOT NULL DEFAULT '[]',
    sso_provider VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (sso_provider IN ('none', 'saml', 'oidc')),
    saml_metadata_url TEXT,
    saml_entity_id TEXT,
    saml_certificate TEXT,
    saml_private_key TEXT,
    use_custom_branding BOOLEAN NOT NULL DEFAULT false,
    email_from_name VARCHAR(100) NOT NULL DEFAULT 'FunctionFly',
    email_from_address VARCHAR(255) NOT NULL DEFAULT 'noreply@functionfly.com',
    require_email_verification BOOLEAN NOT NULL DEFAULT true,
    allow_password_login BOOLEAN NOT NULL DEFAULT true,
    allow_magic_link BOOLEAN NOT NULL DEFAULT true,
    max_login_attempts INT NOT NULL DEFAULT 5,
    lockout_duration_minutes INT NOT NULL DEFAULT 15,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fast tenant lookups
CREATE INDEX IF NOT EXISTS idx_tenant_auth_settings_tenant ON tenant_auth_settings(tenant_id);

-- OAuth providers table for per-tenant OAuth credentials
CREATE TABLE IF NOT EXISTS tenant_oauth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('github', 'google', 'microsoft', 'apple', 'gitlab', 'bitbucket')),
    client_id VARCHAR(255) NOT NULL,
    encrypted_client_secret TEXT NOT NULL,
    encrypted_client_secret_iv TEXT,
    encrypted_client_secret_tag TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    callback_url TEXT,
    scopes JSONB NOT NULL DEFAULT '["user:email", "read:user"]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_tenant_oauth_providers_tenant ON tenant_oauth_providers(tenant_id);

-- Tenant invite codes for user invitation flow
CREATE TABLE IF NOT EXISTS tenant_invite_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'team_member' CHECK (role IN ('team_owner', 'team_admin', 'team_member', 'team_viewer')),
    invited_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    accepted_by UUID REFERENCES users(id),
    max_uses INT NOT NULL DEFAULT 1,
    uses INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_invite_codes_tenant ON tenant_invite_codes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_invite_codes_code ON tenant_invite_codes(code);
CREATE INDEX IF NOT EXISTS idx_tenant_invite_codes_expires ON tenant_invite_codes(expires_at) WHERE accepted_at IS NULL;

-- Tenant user memberships (for team management)
CREATE TABLE IF NOT EXISTS tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'team_member' CHECK (role IN ('team_owner', 'team_admin', 'team_member', 'team_viewer')),
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'invited')),
    UNIQUE(tenant_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_memberships_tenant ON tenant_memberships(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_memberships_user ON tenant_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_memberships_role ON tenant_memberships(role);

-- Audit log for auth events per tenant
CREATE TABLE IF NOT EXISTS tenant_auth_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_auth_audit_tenant ON tenant_auth_audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_auth_audit_user ON tenant_auth_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_auth_audit_created ON tenant_auth_audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_tenant_auth_audit_action ON tenant_auth_audit_log(action);

COMMENT ON TABLE tenant_auth_settings IS 'Per-tenant authentication and authorization configuration for Backend-in-a-Box bundles';
COMMENT ON TABLE tenant_oauth_providers IS 'OAuth provider credentials stored per tenant for social login';
COMMENT ON TABLE tenant_invite_codes IS 'Invite codes for team member invitations';
COMMENT ON TABLE tenant_memberships IS 'Team member roles and membership status per tenant';
COMMENT ON TABLE tenant_auth_audit_log IS 'Audit trail for authentication and authorization events';