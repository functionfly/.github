DROP INDEX IF EXISTS idx_registry_function_versions_sandbox_tier;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS sandbox_tier;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS sandbox_config;
ALTER TABLE registry_functions DROP COLUMN IF EXISTS default_sandbox_tier;
