-- Drop indexes
DROP INDEX IF EXISTS idx_saml_configs_tenant_id;
DROP INDEX IF EXISTS idx_saml_configs_enabled;
DROP INDEX IF EXISTS idx_saml_sessions_tenant_id;
DROP INDEX IF EXISTS idx_saml_sessions_user_id;
DROP INDEX IF EXISTS idx_saml_sessions_not_on_or_after;
DROP INDEX IF EXISTS idx_saml_sessions_saml_name_id;

-- Drop unique constraint
ALTER TABLE saml_configs DROP CONSTRAINT IF EXISTS uq_saml_configs_tenant;

-- Drop tables
DROP TABLE IF EXISTS saml_sessions;
DROP TABLE IF EXISTS saml_configs;
