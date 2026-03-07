-- Create SAML configuration table
CREATE TABLE saml_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    enabled BOOLEAN DEFAULT false,
    idp_metadata XML,
    idp_entity_id VARCHAR(500),
    idp_sso_url VARCHAR(500),
    idp_certificate TEXT,
    sp_entity_id VARCHAR(500) DEFAULT 'functionfly',
    sp_acs_url VARCHAR(500),
    sp_metadata_url VARCHAR(500),
    name_id_format VARCHAR(100) DEFAULT 'emailAddress',
    authn_contexts TEXT[],
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create SAML sessions table
CREATE TABLE saml_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    saml_name_id VARCHAR(255) NOT NULL,
    session_index VARCHAR(255) NOT NULL,
    not_on_or_after TIMESTAMP NOT NULL,
    attributes JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for SAML configs
CREATE INDEX IF NOT EXISTS idx_saml_configs_tenant_id ON saml_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_saml_configs_enabled ON saml_configs(enabled) WHERE enabled = true;

-- Create indexes for SAML sessions
CREATE INDEX IF NOT EXISTS idx_saml_sessions_tenant_id ON saml_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_saml_sessions_user_id ON saml_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_saml_sessions_not_on_or_after ON saml_sessions(not_on_or_after);
CREATE INDEX IF NOT EXISTS idx_saml_sessions_saml_name_id ON saml_sessions(saml_name_id);

-- Add unique constraint for tenant SAML config (one config per tenant)
ALTER TABLE saml_configs ADD CONSTRAINT uq_saml_configs_tenant UNIQUE (tenant_id);
