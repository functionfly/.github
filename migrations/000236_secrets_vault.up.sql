-- Migration: 000077_secrets_vault
-- Description: Zero-knowledge encrypted secrets vault with audit logging and access tokens
-- Created: 2026-03-03
-- Author: FunctionFly

-- =====================================================
-- Table: secrets_vault
-- Description: Stores zero-knowledge encrypted secrets
-- =====================================================
CREATE TABLE IF NOT EXISTS secrets_vault (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    secret_type VARCHAR(50) NOT NULL CHECK (secret_type IN ('api_key', 'oauth_token', 'password', 'certificate')),
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

-- Add comments for secrets_vault table and columns
COMMENT ON TABLE secrets_vault IS 'Zero-knowledge encrypted secrets storage for tenant users';
COMMENT ON COLUMN secrets_vault.id IS 'Unique identifier for the secret';
COMMENT ON COLUMN secrets_vault.tenant_id IS 'Reference to the tenant owning this secret';
COMMENT ON COLUMN secrets_vault.user_id IS 'Reference to the user who created/owns this secret';
COMMENT ON COLUMN secrets_vault.name IS 'Human-readable name for the secret (unique per tenant)';
COMMENT ON COLUMN secrets_vault.description IS 'Optional description of the secret purpose';
COMMENT ON COLUMN secrets_vault.secret_type IS 'Type of secret: api_key, oauth_token, password, or certificate';
COMMENT ON COLUMN secrets_vault.encrypted_value IS 'AES-256-GCM encrypted secret value (server cannot decrypt)';
COMMENT ON COLUMN secrets_vault.encryption_iv IS 'Initialization vector used for encryption';
COMMENT ON COLUMN secrets_vault.encryption_salt IS 'Salt value used for key derivation';
COMMENT ON COLUMN secrets_vault.encryption_auth_tag IS 'GCM authentication tag for integrity verification';
COMMENT ON COLUMN secrets_vault.key_version IS 'Version of encryption key for rotation support';
COMMENT ON COLUMN secrets_vault.metadata IS 'JSON metadata for extensibility (key-value pairs)';
COMMENT ON COLUMN secrets_vault.scopes IS 'JSON array of permission scopes for this secret';
COMMENT ON COLUMN secrets_vault.is_active IS 'Whether the secret is active and usable';
COMMENT ON COLUMN secrets_vault.expires_at IS 'Optional expiration timestamp for the secret';
COMMENT ON COLUMN secrets_vault.created_at IS 'Timestamp when the secret was created';
COMMENT ON COLUMN secrets_vault.updated_at IS 'Timestamp when the secret was last modified';
COMMENT ON COLUMN secrets_vault.last_accessed_at IS 'Timestamp when the secret was last accessed';

-- =====================================================
-- Table: secret_access_tokens
-- Description: Short-lived tokens for accessing secrets
-- =====================================================
CREATE TABLE IF NOT EXISTS secret_access_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets_vault(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    scope VARCHAR(100) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    is_revoked BOOLEAN DEFAULT false
);

-- Add comments for secret_access_tokens table and columns
COMMENT ON TABLE secret_access_tokens IS 'Short-lived access tokens for secret usage with scoped permissions';
COMMENT ON COLUMN secret_access_tokens.id IS 'Unique identifier for the access token';
COMMENT ON COLUMN secret_access_tokens.secret_id IS 'Reference to the secret this token can access';
COMMENT ON COLUMN secret_access_tokens.tenant_id IS 'Reference to the tenant for access control';
COMMENT ON COLUMN secret_access_tokens.token_hash IS 'SHA-256 hash of the token value (token itself is ephemeral)';
COMMENT ON COLUMN secret_access_tokens.scope IS 'Permission scope granted by this token';
COMMENT ON COLUMN secret_access_tokens.expires_at IS 'Token expiration timestamp (required)';
COMMENT ON COLUMN secret_access_tokens.created_at IS 'Timestamp when the token was created';
COMMENT ON COLUMN secret_access_tokens.last_used_at IS 'Timestamp when the token was last used';
COMMENT ON COLUMN secret_access_tokens.ip_address IS 'IP address of the client that created this token';
COMMENT ON COLUMN secret_access_tokens.user_agent IS 'User agent of the client that created this token';
COMMENT ON COLUMN secret_access_tokens.is_revoked IS 'Whether the token has been explicitly revoked';

-- =====================================================
-- Table: secrets_audit_log
-- Description: Audit trail for all secret operations
-- =====================================================
CREATE TABLE IF NOT EXISTS secrets_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    secret_id UUID REFERENCES secrets_vault(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL CHECK (action IN ('create', 'read', 'update', 'delete', 'use', 'revoke')),
    scope VARCHAR(100),
    ip_address VARCHAR(45),
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Add comments for secrets_audit_log table and columns
COMMENT ON TABLE secrets_audit_log IS 'Comprehensive audit trail for all secret-related operations';
COMMENT ON COLUMN secrets_audit_log.id IS 'Unique identifier for the audit log entry';
COMMENT ON COLUMN secrets_audit_log.tenant_id IS 'Reference to the tenant where the action occurred';
COMMENT ON COLUMN secrets_audit_log.secret_id IS 'Reference to the affected secret (NULL if secret was deleted)';
COMMENT ON COLUMN secrets_audit_log.user_id IS 'Reference to the user who performed the action';
COMMENT ON COLUMN secrets_audit_log.action IS 'Type of action: create, read, update, delete, use, or revoke';
COMMENT ON COLUMN secrets_audit_log.scope IS 'Scope associated with the action (if applicable)';
COMMENT ON COLUMN secrets_audit_log.ip_address IS 'IP address of the client that performed the action';
COMMENT ON COLUMN secrets_audit_log.user_agent IS 'User agent of the client that performed the action';
COMMENT ON COLUMN secrets_audit_log.timestamp IS 'Timestamp when the action occurred';
COMMENT ON COLUMN secrets_audit_log.metadata IS 'JSON metadata with additional context about the action';

-- =====================================================
-- Indexes for secrets_vault
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_secrets_vault_tenant ON secrets_vault(tenant_id);
CREATE INDEX IF NOT EXISTS idx_secrets_vault_tenant_name ON secrets_vault(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_secrets_vault_active ON secrets_vault(tenant_id, is_active) WHERE is_active = true;

-- =====================================================
-- Indexes for secret_access_tokens
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_secret ON secret_access_tokens(secret_id);
CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_expires ON secret_access_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_hash ON secret_access_tokens(token_hash);
-- Partial index: only non-revoked tokens (cannot use expires_at > NOW() in predicate - not IMMUTABLE)
CREATE INDEX IF NOT EXISTS idx_secret_access_tokens_active ON secret_access_tokens(secret_id, is_revoked, expires_at) WHERE is_revoked = false;

-- =====================================================
-- Indexes for secrets_audit_log
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_secrets_audit_log_secret ON secrets_audit_log(secret_id);
CREATE INDEX IF NOT EXISTS idx_secrets_audit_log_timestamp ON secrets_audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_secrets_audit_log_tenant ON secrets_audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_secrets_audit_log_tenant_time ON secrets_audit_log(tenant_id, timestamp DESC);
