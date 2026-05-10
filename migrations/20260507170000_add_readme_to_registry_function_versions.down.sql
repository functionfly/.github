DROP INDEX IF EXISTS idx_registry_function_versions_has_readme;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS readme;