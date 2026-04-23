-- Migration: secret_versions_table
-- Description: Add secret versioning and history table for tracking changes over time
-- Created: 2026-04-19
-- Author: FunctionFly

-- =====================================================
-- Table: secret_versions
-- Description: Version history for encrypted secrets
-- =====================================================
CREATE TABLE IF NOT EXISTS secret_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets_vault(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    
    -- Snapshot of secret data at this version
    name VARCHAR(255) NOT NULL,
    description TEXT,
    secret_type VARCHAR(50) NOT NULL CHECK (secret_type IN ('api_key', 'oauth_token', 'password', 'certificate')),
    
    -- Encrypted data snapshot (server never sees plaintext)
    encrypted_value BYTEA NOT NULL,
    encryption_iv BYTEA NOT NULL,
    encryption_salt BYTEA NOT NULL,
    encryption_auth_tag BYTEA NOT NULL,
    key_version INTEGER NOT NULL DEFAULT 1,
    
    -- Scopes and metadata at this version
    scopes JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    
    -- Change tracking
    change_type VARCHAR(20) NOT NULL CHECK (change_type IN ('create', 'update', 'rollback')),
    change_summary TEXT, -- Brief description of what changed (e.g., "Updated API key value")
    
    -- Actor information (who made this version)
    actor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_type VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (actor_type IN ('user', 'system', 'api_key')),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure version numbers are unique per secret and sequential
    CONSTRAINT unique_secret_version UNIQUE (secret_id, version_number)
);

-- Add comments for secret_versions table and columns
COMMENT ON TABLE secret_versions IS 'Version history for secrets, enabling rollback and audit trails';
COMMENT ON COLUMN secret_versions.id IS 'Unique identifier for this version record';
COMMENT ON COLUMN secret_versions.secret_id IS 'Reference to the secret this version belongs to';
COMMENT ON COLUMN secret_versions.tenant_id IS 'Reference to the tenant for security filtering';
COMMENT ON COLUMN secret_versions.version_number IS 'Sequential version number (1, 2, 3, ...) per secret';
COMMENT ON COLUMN secret_versions.name IS 'Name of the secret at this version';
COMMENT ON COLUMN secret_versions.description IS 'Description of the secret at this version';
COMMENT ON COLUMN secret_versions.secret_type IS 'Type of secret at this version';
COMMENT ON COLUMN secret_versions.encrypted_value IS 'AES-256-GCM encrypted value at this version';
COMMENT ON COLUMN secret_versions.encryption_iv IS 'Initialization vector used for encryption at this version';
COMMENT ON COLUMN secret_versions.encryption_salt IS 'Salt value used for key derivation at this version';
COMMENT ON COLUMN secret_versions.encryption_auth_tag IS 'GCM authentication tag for integrity at this version';
COMMENT ON COLUMN secret_versions.key_version IS 'Encryption key version at this version';
COMMENT ON COLUMN secret_versions.scopes IS 'JSON array of permission scopes at this version';
COMMENT ON COLUMN secret_versions.metadata IS 'JSON metadata at this version';
COMMENT ON COLUMN secret_versions.change_type IS 'Type of change: create, update, or rollback';
COMMENT ON COLUMN secret_versions.change_summary IS 'Human-readable summary of what changed';
COMMENT ON COLUMN secret_versions.actor_id IS 'User who created this version';
COMMENT ON COLUMN secret_versions.actor_type IS 'Type of actor who created this version';
COMMENT ON COLUMN secret_versions.created_at IS 'Timestamp when this version was created';

-- =====================================================
-- Indexes for secret_versions
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_secret_versions_secret ON secret_versions(secret_id);
CREATE INDEX IF NOT EXISTS idx_secret_versions_tenant ON secret_versions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_secret_versions_created_at ON secret_versions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_secret_versions_secret_version ON secret_versions(secret_id, version_number DESC);

-- =====================================================
-- Function: get_next_secret_version
-- Description: Returns the next version number for a secret
-- =====================================================
CREATE OR REPLACE FUNCTION get_next_secret_version(p_secret_id UUID)
RETURNS INTEGER AS $$
DECLARE
    next_version INTEGER;
BEGIN
    SELECT COALESCE(MAX(version_number), 0) + 1
    INTO next_version
    FROM secret_versions
    WHERE secret_id = p_secret_id;
    
    RETURN next_version;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- Function: get_secret_current_version
-- Description: Returns the current version number for a secret
-- =====================================================
CREATE OR REPLACE FUNCTION get_secret_current_version(p_secret_id UUID)
RETURNS INTEGER AS $$
DECLARE
    current_version INTEGER;
BEGIN
    SELECT COALESCE(MAX(version_number), 0)
    INTO current_version
    FROM secret_versions
    WHERE secret_id = p_secret_id;
    
    RETURN current_version;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- Trigger Function: auto_create_initial_version
-- Description: Automatically creates a version 1 record when a secret is created
-- =====================================================
CREATE OR REPLACE FUNCTION auto_create_initial_version()
RETURNS TRIGGER AS $$
BEGIN
    -- Create initial version record for the new secret
    INSERT INTO secret_versions (
        secret_id,
        tenant_id,
        version_number,
        name,
        description,
        secret_type,
        encrypted_value,
        encryption_iv,
        encryption_salt,
        encryption_auth_tag,
        key_version,
        scopes,
        metadata,
        change_type,
        change_summary,
        actor_id,
        actor_type
    ) VALUES (
        NEW.id,
        NEW.tenant_id,
        1,
        NEW.name,
        NEW.description,
        NEW.secret_type,
        NEW.encrypted_value,
        NEW.encryption_iv,
        NEW.encryption_salt,
        NEW.encryption_auth_tag,
        NEW.key_version,
        COALESCE(NEW.scopes, '[]'::jsonb),
        COALESCE(NEW.metadata, '{}'::jsonb),
        'create',
        'Initial secret creation',
        NEW.user_id,
        'user'
    );
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- Trigger: trigger_secret_version_on_create
-- Description: Fires after insert on secrets_vault to create initial version
-- =====================================================
DROP TRIGGER IF EXISTS trigger_secret_version_on_create ON secrets_vault;
CREATE TRIGGER trigger_secret_version_on_create
    AFTER INSERT ON secrets_vault
    FOR EACH ROW
    EXECUTE FUNCTION auto_create_initial_version();

-- =====================================================
-- Add version tracking columns to secrets_vault
-- =====================================================
ALTER TABLE secrets_vault 
    ADD COLUMN IF NOT EXISTS current_version INTEGER DEFAULT 1,
    ADD COLUMN IF NOT EXISTS last_modified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS last_modified_at TIMESTAMP WITH TIME ZONE;

-- Add comments for new columns
COMMENT ON COLUMN secrets_vault.current_version IS 'Current version number from secret_versions table';
COMMENT ON COLUMN secrets_vault.last_modified_by IS 'User who last modified this secret';
COMMENT ON COLUMN secrets_vault.last_modified_at IS 'Timestamp of last modification (not access)';

-- Update existing secrets to set current_version = 1 if not set
UPDATE secrets_vault 
SET current_version = 1 
WHERE current_version IS NULL;

-- Create initial versions for existing secrets that don't have them yet
INSERT INTO secret_versions (
    secret_id,
    tenant_id,
    version_number,
    name,
    description,
    secret_type,
    encrypted_value,
    encryption_iv,
    encryption_salt,
    encryption_auth_tag,
    key_version,
    scopes,
    metadata,
    change_type,
    change_summary,
    actor_id,
    actor_type
)
SELECT 
    s.id,
    s.tenant_id,
    1,
    s.name,
    s.description,
    s.secret_type,
    s.encrypted_value,
    s.encryption_iv,
    s.encryption_salt,
    s.encryption_auth_tag,
    s.key_version,
    COALESCE(s.scopes, '[]'::jsonb),
    COALESCE(s.metadata, '{}'::jsonb),
    'create',
    'Initial version (migration)',
    s.user_id,
    'user'
FROM secrets_vault s
LEFT JOIN secret_versions sv ON s.id = sv.secret_id
WHERE sv.id IS NULL;
