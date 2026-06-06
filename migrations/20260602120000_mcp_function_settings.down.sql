-- 20260602120000_mcp_function_settings.down.sql
BEGIN;
DROP TRIGGER IF EXISTS trg_mcp_settings_updated_at ON registry_function_mcp_settings;
DROP FUNCTION IF EXISTS trg_mcp_settings_set_updated_at();
DROP TABLE IF EXISTS registry_function_mcp_settings CASCADE;
COMMIT;
