-- +migrate Up
CREATE TABLE IF NOT EXISTS tenant_auth_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    mfa_required BOOLEAN NOT NULL DEFAULT false,
    mfa_mode TEXT NOT NULL DEFAULT 'optional',
    password_policy JSONB NOT NULL DEFAULT '{"min_length":8,"require_uppercase":true,"require_lowercase":true,"require_digit":true,"require_special":true}',
    session_timeout_minutes INTEGER NOT NULL DEFAULT 480,
    ip_allowlist_enabled BOOLEAN NOT NULL DEFAULT false,
    ip_allowlist TEXT,
    allowed_domains TEXT,
    sso_provider TEXT,
    saml_metadata_url TEXT,
    saml_entity_id TEXT,
    saml_certificate TEXT,
    allow_password_login BOOLEAN NOT NULL DEFAULT true,
    allow_magic_link BOOLEAN NOT NULL DEFAULT true,
    require_email_verification BOOLEAN NOT NULL DEFAULT true,
    lockout_duration_minutes INTEGER NOT NULL DEFAULT 15,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tenant_auth_settings_tenant_id ON tenant_auth_settings(tenant_id);

-- +migrate Down
DROP TABLE IF EXISTS tenant_auth_settings;