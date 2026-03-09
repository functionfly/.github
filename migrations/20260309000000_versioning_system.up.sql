-- Versioning System Phase 1: Foundation
-- Creates API version management and function version tracking tables

-- 1. API Versions table - Platform API lifecycle management
CREATE TABLE IF NOT EXISTS api_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version VARCHAR(20) NOT NULL UNIQUE,
    path_prefix VARCHAR(50) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deprecated', 'sunset', 'archived')),

    -- Release metadata
    released_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deprecated_at TIMESTAMPTZ,
    sunset_at TIMESTAMPTZ,
    sunset_message TEXT,

    -- Additional metadata
    metadata JSONB DEFAULT '{}'::jsonb,
    openapi_spec_url TEXT,
    changelog_url TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for API versions
CREATE INDEX IF NOT EXISTS idx_api_versions_status ON api_versions(status);
CREATE INDEX IF NOT EXISTS idx_api_versions_released_at ON api_versions(released_at);

-- Insert default API versions
INSERT INTO api_versions (version, path_prefix, status, released_at, metadata)
VALUES
    ('v1', '/v1', 'active', '2024-01-01T00:00:00Z'::timestamptz, '{"features": ["basic_functions", "deployments", "registry"]}'::jsonb),
    ('v2', '/v2', 'active', '2025-01-15T00:00:00Z'::timestamptz, '{"features": ["graphql", "webhooks", "advanced_deployments"]}'::jsonb)
ON CONFLICT (version) DO NOTHING;

-- 2. Function versions table - Extends registry_function_versions with version state
-- First, add columns if they don't exist (for compatibility with existing registry)
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS version_state VARCHAR(20) DEFAULT 'published' CHECK (version_state IN ('draft', 'published', 'deprecated', 'archived'));
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS deprecation_reason TEXT;
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS replaced_by_version VARCHAR(50);
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS migration_guide TEXT;
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS deprecated_at TIMESTAMPTZ;

-- Add indexes for new columns
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_state ON registry_function_versions(version_state) WHERE version_state IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_deprecated ON registry_function_versions(deprecated_at) WHERE deprecated_at IS NOT NULL;

-- 3. Function version changelog table - Track changes between versions
CREATE TABLE IF NOT EXISTS function_version_changelog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,

    -- Change details
    change_type VARCHAR(20) NOT NULL CHECK (change_type IN ('major', 'minor', 'patch', 'breaking', 'feature', 'fix', 'security', 'deprecation')),
    change_category VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,

    -- Additional change metadata
    breaking_changes JSONB DEFAULT '[]'::jsonb,
    migration_steps JSONB DEFAULT '[]'::jsonb,

    -- Authorship
    created_by UUID REFERENCES users(id),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_function_version_changelog UNIQUE (function_id, version)
);

-- Indexes for changelog
CREATE INDEX IF NOT EXISTS idx_function_version_changelog_function_id ON function_version_changelog(function_id);
CREATE INDEX IF NOT EXISTS idx_function_version_changelog_version ON function_version_changelog(version);
CREATE INDEX IF NOT EXISTS idx_function_version_changelog_change_type ON function_version_changelog(change_type);
CREATE INDEX IF NOT EXISTS idx_function_version_changelog_created_at ON function_version_changelog(created_at);

-- 4. Deployment versions table - Track deployments per version
CREATE TABLE IF NOT EXISTS deployment_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_version VARCHAR(50) NOT NULL,

    -- Deployment reference
    deployment_id UUID REFERENCES deployments(id) ON DELETE SET NULL,

    -- Provider details
    provider VARCHAR(50) NOT NULL,
    region VARCHAR(50),

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'building', 'deploying', 'success', 'failed', 'rolled_back')),

    -- Artifact info
    artifact_uri TEXT,
    checksum VARCHAR(64),

    -- Rollback support
    rollback_id UUID REFERENCES deployment_versions(id),

    -- Metadata
    metadata JSONB DEFAULT '{}'::jsonb,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Indexes for deployment versions
CREATE INDEX IF NOT EXISTS idx_deployment_versions_function_id ON deployment_versions(function_id);
CREATE INDEX IF NOT EXISTS idx_deployment_versions_function_version ON deployment_versions(function_version);
CREATE INDEX IF NOT EXISTS idx_deployment_versions_status ON deployment_versions(status);
CREATE INDEX IF NOT EXISTS idx_deployment_versions_provider ON deployment_versions(provider);
CREATE INDEX IF NOT EXISTS idx_deployment_versions_created_at ON deployment_versions(created_at);

-- 5. Service contracts table - Internal service communication versioning
CREATE TABLE IF NOT EXISTS service_contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(100) NOT NULL,
    contract_version VARCHAR(50) NOT NULL,
    contract_type VARCHAR(20) NOT NULL CHECK (contract_type IN ('grpc', 'http', 'event', 'message')),

    -- Schema
    schema JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'deprecated', 'removed')),

    -- Release tracking
    introduced_in_release VARCHAR(50),
    deprecated_in_release VARCHAR(50),
    removed_in_release VARCHAR(50),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT uq_service_contract UNIQUE (service_name, contract_version, contract_type)
);

-- Indexes for service contracts
CREATE INDEX IF NOT EXISTS idx_service_contracts_service_name ON service_contracts(service_name);
CREATE INDEX IF NOT EXISTS idx_service_contracts_status ON service_contracts(status);
