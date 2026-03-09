-- Rollback Versioning System Phase 1

-- Drop service contracts table
DROP TABLE IF EXISTS service_contracts CASCADE;

-- Drop deployment versions table
DROP TABLE IF EXISTS deployment_versions CASCADE;

-- Drop function version changelog table
DROP TABLE IF EXISTS function_version_changelog CASCADE;

-- Remove indexes from registry_function_versions
DROP INDEX IF EXISTS idx_registry_function_versions_state;
DROP INDEX IF EXISTS idx_registry_function_versions_deprecated;

-- Remove added columns from registry_function_versions
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS migration_guide;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS replaced_by_version;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS deprecation_reason;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS version_state;

-- Drop API versions table
DROP TABLE IF EXISTS api_versions CASCADE;
