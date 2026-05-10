-- Migration: 20260508153000_unify_functions_and_registry
-- Description: Unify functions table with registry_functions by adding missing tenant-specific columns
-- This enables a single API surface for both public registry functions and tenant-private functions

BEGIN;

-- Add missing columns to registry_functions for tenant function support
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS providers TEXT[] DEFAULT '{}';
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS region VARCHAR(100) DEFAULT 'global';
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS code TEXT;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS env_vars JSONB DEFAULT '[]'::jsonb;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS schedule JSONB;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS playground_enabled BOOLEAN DEFAULT false;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS playground_config JSONB DEFAULT '{}'::jsonb;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'draft';
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS app_id UUID REFERENCES apps(id);

-- Existing registry_functions already has:
-- - tenant_id (UUID, nullable, references tenants)
-- - owner_user_id (UUID, nullable, references users)
-- These are kept for ownership tracking.

-- Add index for tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_registry_functions_tenant_id ON registry_functions(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_registry_functions_app_id ON registry_functions(app_id) WHERE app_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_registry_functions_status ON registry_functions(status);

-- Create non-unique index for tenant-scoped function name lookups
-- Note: We intentionally do NOT create a unique constraint here because
-- the migration revealed duplicate (tenant_id, name) pairs exist in production.
-- Application-level enforcement of uniqueness is recommended.
CREATE INDEX IF NOT EXISTS idx_registry_functions_tenant_name ON registry_functions(tenant_id, name) WHERE tenant_id IS NOT NULL;

-- Backfill existing "functionfly" authored functions with defaults for tenant-specific fields
UPDATE registry_functions SET
    providers = '{}',
    region = 'global',
    code = COALESCE(code, ''),
    env_vars = COALESCE(env_vars, '[]'::jsonb),
    playground_enabled = COALESCE(playground_enabled, false),
    playground_config = COALESCE(playground_config, '{}'::jsonb),
    status = 'deployed'
WHERE author = 'functionfly' AND status = 'draft';

-- Function to check if a function is tenant-private
CREATE OR REPLACE FUNCTION is_tenant_function(f registry_functions) RETURNS BOOLEAN AS $$
BEGIN
    RETURN f.tenant_id IS NOT NULL;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to get function address (author/name for public, tenant/name for private)
CREATE OR REPLACE FUNCTION get_function_address(f registry_functions) RETURNS TEXT AS $$
BEGIN
    IF f.tenant_id IS NOT NULL THEN
        -- For tenant functions, use tenant_id as author
        RETURN COALESCE(f.tenant_id::text, 'unknown') || '/' || f.name;
    ELSE
        -- For public functions
        RETURN f.author || '/' || f.name;
    END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMIT;