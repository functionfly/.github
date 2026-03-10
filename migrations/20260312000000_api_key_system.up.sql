-- Migration: 20260312000000_api_key_system
-- Description: API Key System - Core tables for API key management
-- Created: 2026-03-12
-- Author: FunctionFly

-- =====================================================
-- Table: api_keys
-- Description: Main API keys table with key management features
-- =====================================================
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    key_type VARCHAR(50) NOT NULL CHECK (key_type IN ('platform', 'function', 'agent', 'environment', 'oauth')),
    key_prefix VARCHAR(10) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_version INTEGER NOT NULL DEFAULT 1,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_rotated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    rotation_frequency_days INTEGER DEFAULT 90,
    rate_limit_rpm INTEGER DEFAULT 1000,
    rate_limit_rph INTEGER DEFAULT 60000,
    rate_limit_rpd INTEGER DEFAULT 1000000,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT tenant_api_keys_unique UNIQUE (tenant_id, name)
);

-- Add comments for api_keys table and columns
COMMENT ON TABLE api_keys IS 'API keys for tenant access with key management, rate limiting, and rotation support';
COMMENT ON COLUMN api_keys.id IS 'Unique identifier for the API key';
COMMENT ON COLUMN api_keys.tenant_id IS 'Reference to the tenant owning this API key';
COMMENT ON COLUMN api_keys.user_id IS 'Reference to the user who created/owns this API key';
COMMENT ON COLUMN api_keys.name IS 'Human-readable name for the API key (unique per tenant)';
COMMENT ON COLUMN api_keys.description IS 'Optional description of the API key purpose';
COMMENT ON COLUMN api_keys.key_type IS 'Type of API key: platform, function, agent, environment, or oauth';
COMMENT ON COLUMN api_keys.key_prefix IS 'Key type prefix (ffp_, fff_, aep_, ffe_, ffo_)';
COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 hash of the API key for secure storage';
COMMENT ON COLUMN api_keys.key_version IS 'Version number for key rotation support';
COMMENT ON COLUMN api_keys.expires_at IS 'Optional expiration timestamp for the API key';
COMMENT ON COLUMN api_keys.last_rotated_at IS 'Timestamp when the key was last rotated';
COMMENT ON COLUMN api_keys.rotation_frequency_days IS 'Number of days between automatic rotations';
COMMENT ON COLUMN api_keys.rate_limit_rpm IS 'Requests per minute limit';
COMMENT ON COLUMN api_keys.rate_limit_rph IS 'Requests per hour limit';
COMMENT ON COLUMN api_keys.rate_limit_rpd IS 'Requests per day limit';
COMMENT ON COLUMN api_keys.is_active IS 'Whether the API key is active and usable';
COMMENT ON COLUMN api_keys.metadata IS 'JSON metadata for extensibility';
COMMENT ON COLUMN api_keys.created_at IS 'Timestamp when the API key was created';
COMMENT ON COLUMN api_keys.updated_at IS 'Timestamp when the API key was last modified';
COMMENT ON COLUMN api_keys.last_used_at IS 'Timestamp when the API key was last used';

-- =====================================================
-- Table: api_key_rotations
-- Description: History of API key rotations
-- =====================================================
CREATE TABLE IF NOT EXISTS api_key_rotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    rotated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id),
    key_hash VARCHAR(255) NOT NULL,
    rotation_reason VARCHAR(50) DEFAULT 'manual' CHECK (rotation_reason IN ('manual', 'automatic', 'compromised')),
    metadata JSONB DEFAULT '{}'
);

-- Add comments for api_key_rotations table and columns
COMMENT ON TABLE api_key_rotations IS 'Audit trail of all API key rotations';
COMMENT ON COLUMN api_key_rotations.id IS 'Unique identifier for the rotation record';
COMMENT ON COLUMN api_key_rotations.api_key_id IS 'Reference to the API key that was rotated';
COMMENT ON COLUMN api_key_rotations.rotated_at IS 'Timestamp when the rotation occurred';
COMMENT ON COLUMN api_key_rotations.expires_at IS 'Expiration timestamp for the rotated key';
COMMENT ON COLUMN api_key_rotations.created_by IS 'User who initiated the rotation';
COMMENT ON COLUMN api_key_rotations.key_hash IS 'Hash of the key at time of rotation';
COMMENT ON COLUMN api_key_rotations.rotation_reason IS 'Reason for rotation: manual, automatic, or compromised';
COMMENT ON COLUMN api_key_rotations.metadata IS 'JSON metadata with additional context';

-- =====================================================
-- Table: api_key_permissions
-- Description: Granular permissions for API keys
-- =====================================================
CREATE TABLE IF NOT EXISTS api_key_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    permission VARCHAR(50) NOT NULL CHECK (permission IN ('read', 'write', 'execute', 'admin')),
    resource_type VARCHAR(50) NOT NULL CHECK (resource_type IN ('function', 'app', 'tenant', 'registry', 'deployment', 'secret')),
    resource_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(api_key_id, resource_type, resource_id, permission)
);

-- Add comments for api_key_permissions table and columns
COMMENT ON TABLE api_key_permissions IS 'Granular permissions assigned to API keys';
COMMENT ON COLUMN api_key_permissions.id IS 'Unique identifier for the permission';
COMMENT ON COLUMN api_key_permissions.api_key_id IS 'Reference to the API key';
COMMENT ON COLUMN api_key_permissions.permission IS 'Permission level: read, write, execute, or admin';
COMMENT ON COLUMN api_key_permissions.resource_type IS 'Type of resource being accessed';
COMMENT ON COLUMN api_key_permissions.resource_id IS 'ID of the specific resource';
COMMENT ON COLUMN api_key_permissions.created_at IS 'Timestamp when the permission was created';

-- =====================================================
-- Table: api_key_environments
-- Description: Environment mapping for API keys
-- =====================================================
CREATE TABLE IF NOT EXISTS api_key_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL,
    environment_name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(api_key_id, environment_id)
);

-- Add comments for api_key_environments table and columns
COMMENT ON TABLE api_key_environments IS 'Environment associations for API keys';
COMMENT ON COLUMN api_key_environments.id IS 'Unique identifier for the environment association';
COMMENT ON COLUMN api_key_environments.api_key_id IS 'Reference to the API key';
COMMENT ON COLUMN api_key_environments.environment_id IS 'Reference to the environment';
COMMENT ON COLUMN api_key_environments.environment_name IS 'Name of the environment';
COMMENT ON COLUMN api_key_environments.created_at IS 'Timestamp when the environment was associated';

-- =====================================================
-- Indexes for api_keys
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant ON api_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_name ON api_keys(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(tenant_id, is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_api_keys_expires ON api_keys(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_type ON api_keys(key_type);

-- =====================================================
-- Indexes for api_key_rotations
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_api_key_rotations_key_id ON api_key_rotations(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_rotations_rotated_at ON api_key_rotations(rotated_at);

-- =====================================================
-- Indexes for api_key_permissions
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_api_key_permissions_key_id ON api_key_permissions(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_permissions_resource ON api_key_permissions(resource_type, resource_id);

-- =====================================================
-- Indexes for api_key_environments
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_api_key_environments_key_id ON api_key_environments(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_environments_env_id ON api_key_environments(environment_id);
