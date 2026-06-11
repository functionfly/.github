-- GoBetterAuth SAML and SCIM Tables
-- Creates tables for Enterprise SSO and SCIM provisioning

-- SAML Configuration
CREATE TABLE IF NOT EXISTS gba_saml_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT FALSE,
    idp_entity_id VARCHAR(500) NOT NULL,
    idp_sso_url VARCHAR(500) NOT NULL,
    idp_certificate TEXT NOT NULL,
    sp_entity_id VARCHAR(500) NOT NULL,
    acs_url VARCHAR(500) NOT NULL,
    name_id_format VARCHAR(100) DEFAULT 'emailAddress',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id)
);

-- SAML Sessions for Single Logout support
CREATE TABLE IF NOT EXISTS gba_saml_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES gba_users(id) ON DELETE CASCADE,
    name_id VARCHAR(255) NOT NULL,
    session_index VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- SCIM Configuration
CREATE TABLE IF NOT EXISTS gba_scim_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT FALSE,
    token_hash TEXT NOT NULL,
    sync_groups BOOLEAN DEFAULT TRUE,
    sync_users BOOLEAN DEFAULT TRUE,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id)
);

-- SCIM Users for provisioning
CREATE TABLE IF NOT EXISTS gba_scim_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
    external_id VARCHAR(255),
    user_name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    emails JSONB DEFAULT '[]',
    active BOOLEAN DEFAULT TRUE,
    groups JSONB DEFAULT '[]',
    raw JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, external_id)
);

-- SCIM Groups for provisioning
CREATE TABLE IF NOT EXISTS gba_scim_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES gba_tenants(id) ON DELETE CASCADE,
    external_id VARCHAR(255),
    display_name VARCHAR(255) NOT NULL,
    members JSONB DEFAULT '[]',
    raw JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(tenant_id, external_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_gba_saml_config_tenant ON gba_saml_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gba_saml_config_enabled ON gba_saml_configs(enabled) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_gba_saml_session_user ON gba_saml_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_gba_saml_session_tenant ON gba_saml_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gba_saml_session_expires ON gba_saml_sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_gba_scim_config_tenant ON gba_scim_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gba_scim_config_enabled ON gba_scim_configs(enabled) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_gba_scim_user_tenant ON gba_scim_users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gba_scim_user_external ON gba_scim_users(tenant_id, external_id);
CREATE INDEX IF NOT EXISTS idx_gba_scim_user_username ON gba_scim_users(tenant_id, user_name);
CREATE INDEX IF NOT EXISTS idx_gba_scim_user_active ON gba_scim_users(tenant_id, active);
CREATE INDEX IF NOT EXISTS idx_gba_scim_group_tenant ON gba_scim_groups(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gba_scim_group_external ON gba_scim_groups(tenant_id, external_id);
CREATE INDEX IF NOT EXISTS idx_gba_scim_group_name ON gba_scim_groups(tenant_id, display_name);

-- Add comments for documentation
COMMENT ON TABLE gba_saml_configs IS 'SAML Identity Provider configuration for GoBetterAuth tenants';
COMMENT ON TABLE gba_saml_sessions IS 'Active SAML sessions for Single Logout support';
COMMENT ON TABLE gba_scim_configs IS 'SCIM provisioning configuration for GoBetterAuth tenants';
COMMENT ON TABLE gba_scim_users IS 'SCIM provisioned users';
COMMENT ON TABLE gba_scim_groups IS 'SCIM provisioned groups';

COMMENT ON COLUMN gba_saml_configs.idp_entity_id IS 'Identity Provider Entity ID (e.g., https://login.microsoftonline.com/...)';
COMMENT ON COLUMN gba_saml_configs.idp_sso_url IS 'Identity Provider Single Sign-On URL';
COMMENT ON COLUMN gba_saml_configs.idp_certificate IS 'Identity Provider X.509 certificate (PEM encoded)';
COMMENT ON COLUMN gba_saml_configs.sp_entity_id IS 'Service Provider Entity ID (typically your app URL)';
COMMENT ON COLUMN gba_saml_configs.acs_url IS 'Assertion Consumer Service URL where IdP sends SAML responses';

COMMENT ON COLUMN gba_scim_configs.token_hash IS 'Bearer token hash for SCIM authentication (bcrypt)';
COMMENT ON COLUMN gba_scim_configs.sync_groups IS 'Enable group synchronization via SCIM';
COMMENT ON COLUMN gba_scim_configs.sync_users IS 'Enable user synchronization via SCIM';
