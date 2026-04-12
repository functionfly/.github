-- Drop execution retention settings table and related objects
DROP INDEX IF EXISTS idx_execution_retention_settings_updated_at;
DROP INDEX IF EXISTS idx_execution_retention_settings_single_active;
DROP TABLE IF EXISTS execution_retention_settings;
