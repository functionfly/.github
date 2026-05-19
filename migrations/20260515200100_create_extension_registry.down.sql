-- Drop extension_registry table
DROP INDEX IF EXISTS idx_extension_registry_installed_at;
DROP INDEX IF EXISTS idx_extension_registry_category;
DROP INDEX IF EXISTS idx_extension_registry_status;
DROP INDEX IF EXISTS idx_extension_registry_name;
DROP INDEX IF EXISTS idx_extension_registry_tenant;
DROP TABLE IF EXISTS extension_registry;