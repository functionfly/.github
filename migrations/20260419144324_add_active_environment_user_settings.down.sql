-- Remove active environment support from user settings

DROP INDEX IF EXISTS idx_users_settings_active_env;
DROP INDEX IF EXISTS idx_users_tenant_id_settings_env;
