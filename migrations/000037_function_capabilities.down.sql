-- Rollback: Remove capabilities fields
-- Date: 2024-02-20
-- Description: Removes capabilities columns

-- Drop indexes first
DROP INDEX IF EXISTS idx_registry_function_versions_capabilities;
DROP INDEX IF EXISTS idx_registry_functions_capabilities;
DROP INDEX IF EXISTS idx_function_configs_capabilities;

-- Remove columns
ALTER TABLE registry_function_versions
DROP COLUMN IF EXISTS capabilities;

ALTER TABLE registry_functions
DROP COLUMN IF EXISTS capabilities;

ALTER TABLE function_configs
DROP COLUMN IF EXISTS capabilities;