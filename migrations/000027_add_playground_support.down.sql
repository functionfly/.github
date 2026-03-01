-- Remove playground support and app_id from functions table
DROP INDEX IF EXISTS idx_functions_playground_enabled;
DROP INDEX IF EXISTS idx_functions_app_id;
DROP INDEX IF EXISTS idx_apps_slug;
ALTER TABLE functions DROP COLUMN IF EXISTS playground_config;
ALTER TABLE functions DROP COLUMN IF EXISTS playground_enabled;
ALTER TABLE functions DROP COLUMN IF EXISTS app_id;
