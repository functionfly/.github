-- Migration: Add capabilities field to function_configs
-- Date: 2024-02-20
-- Description: Adds capabilities column for sandbox permission enforcement

-- Add capabilities column to function_configs table
ALTER TABLE function_configs
ADD COLUMN IF NOT EXISTS capabilities TEXT[] DEFAULT '{}';

-- Add index for capability-based queries (e.g., finding all functions with a specific capability)
CREATE INDEX IF NOT EXISTS idx_function_configs_capabilities
ON function_configs USING GIN(capabilities);

-- Add capabilities column to registry_functions table (for public registry)
ALTER TABLE registry_functions
ADD COLUMN IF NOT EXISTS capabilities TEXT[] DEFAULT '{}';

-- Add index for registry capability queries
CREATE INDEX IF NOT EXISTS idx_registry_functions_capabilities
ON registry_functions USING GIN(capabilities);

-- Add capabilities column to registry_function_versions table
ALTER TABLE registry_function_versions
ADD COLUMN IF NOT EXISTS capabilities JSONB DEFAULT '[]';

-- Add index for version capability queries
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_capabilities
ON registry_function_versions USING GIN(capabilities);
