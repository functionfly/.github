-- Secure Secrets Vault
-- Zero-knowledge encrypted secrets storage

-- Secrets vault table
CREATE TABLE IF NOT EXISTS secrets_vault (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    secret_type VARCHAR(50) NOT NULL DEFAULT 'api_key', -- 'api_key', 'oauth_token', 'password', 'certificate', 'generic'
    encrypted_value BYTEA NOT NULL,
    encryption_iv BYTEA NOT NULL,
    encryption_salt BYTEA NOT NULL,
    encryption_auth_tag BYTEA NOT NULL,
    key_version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB DEFAULT '{}',
    scopes JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT tenant_secrets_unique UNIQUE (tenant_id, name)
);

-- Secret access tokens (short-lived)
CREATE TABLE IF NOT EXISTS secret_access_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets_vault(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    token_hash VARCHAR(255) NOT NULL,
    scope VARCHAR(100) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    is_revoked BOOLEAN DEFAULT false
);

-- Audit log for secret access
CREATE TABLE IF NOT EXISTS secrets_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    secret_id UUID REFERENCES secrets_vault(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL, -- 'create', 'read', 'update', 'delete', 'use', 'revoke', 'generate_token'
    scope VARCHAR(100),
    ip_address VARCHAR(45),
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_secrets_vault_tenant ON secrets_vault(tenant_id);
CREATE INDEX IF NOT EXISTS idx_secrets_vault_tenant_name ON secrets_vault(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_secrets_vault_user ON secrets_vault(user_id);
CREATE INDEX IF NOT EXISTS idx_secrets_vault_type ON secrets_vault(secret_type);
CREATE INDEX IF NOT EXISTS idx_secrets_vault_active ON secrets_vault(is_active);

CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_secret ON secret_access_tokens(secret_id);
CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_tenant ON secret_access_tokens(tenant_id);
CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_expires ON secret_access_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_token_hash ON secret_access_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_revoked ON secret_access_tokens(is_revoked);

CREATE INDEX IF NOT EXISTS idx_secrets_audit_log_secret ON secrets_audit_log(secret_id);
CREATE INDEX IF NOT EXISTS idx_secrets_audit_log_timestamp ON secrets_audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_secrets_audit_log_tenant ON secrets_audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_secrets_audit_log_user ON secrets_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_secrets_audit_log_action ON secrets_audit_log(action);

-- Row Level Security Policies
ALTER TABLE secrets_vault ENABLE ROW LEVEL SECURITY;
ALTER TABLE secret_access_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE secrets_audit_log ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy for secrets_vault
CREATE POLICY secrets_vault_tenant_isolation ON secrets_vault
    FOR ALL
    USING (tenant_id IN (SELECT tenant_id FROM users WHERE id = current_setting('app.current_user_id', true)::uuid));

-- Tenant isolation policy for secret_access_tokens
CREATE POLICY secret_access_tokens_tenant_isolation ON secret_access_tokens
    FOR ALL
    USING (tenant_id IN (SELECT tenant_id FROM users WHERE id = current_setting('app.current_user_id', true)::uuid));

-- Tenant isolation policy for secrets_audit_log
CREATE POLICY secrets_audit_log_tenant_isolation ON secrets_audit_log
    FOR ALL
    USING (tenant_id IN (SELECT tenant_id FROM users WHERE id = current_setting('app.current_user_id', true)::uuid));

-- Comments for documentation
COMMENT ON TABLE secrets_vault IS 'Encrypted secrets storage with zero-knowledge architecture';
COMMENT ON TABLE secret_access_tokens IS 'Short-lived access tokens for secret usage';
COMMENT ON TABLE secrets_audit_log IS 'Audit trail for secret operations';
