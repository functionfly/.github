-- Versioning System Phase 2: Publishing, Aliases, and Rollback
-- Migration: 20260310000000_versioning_phase2

-- Create version aliases table
CREATE TABLE IF NOT EXISTS version_aliases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id),
    alias VARCHAR(50) NOT NULL,
    version_id UUID NOT NULL REFERENCES registry_function_versions(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(function_id, alias)
);

CREATE INDEX IF NOT EXISTS idx_version_aliases_function ON version_aliases(function_id);
CREATE INDEX IF NOT EXISTS idx_version_aliases_alias ON version_aliases(alias);

-- Create rollback records table
CREATE TABLE IF NOT EXISTS rollback_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id),
    from_version VARCHAR(50) NOT NULL,
    to_version VARCHAR(50) NOT NULL,
    strategy VARCHAR(20) NOT NULL DEFAULT 'immediate',
    status VARCHAR(20) NOT NULL DEFAULT 'initiated',
    initiated_by UUID REFERENCES users(id),
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_rollback_records_function ON rollback_records(function_id);
CREATE INDEX IF NOT EXISTS idx_rollback_records_initiated_at ON rollback_records(initiated_at DESC);

-- Add sunset_at to function versions for deprecation
ALTER TABLE registry_function_versions
ADD COLUMN IF NOT EXISTS sunset_at TIMESTAMPTZ;

-- Add published_at if not exists
ALTER TABLE registry_function_versions
ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ;

-- Add archived_at if not exists
ALTER TABLE registry_function_versions
ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

-- Add is_default column to api_versions
ALTER TABLE api_versions
ADD COLUMN IF NOT EXISTS is_default BOOLEAN DEFAULT false;
