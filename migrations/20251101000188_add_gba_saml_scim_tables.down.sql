-- Drop GoBetterAuth SAML and SCIM tables

-- Drop indexes first
DROP INDEX IF EXISTS idx_gba_scim_group_name;
DROP INDEX IF EXISTS idx_gba_scim_group_external;
DROP INDEX IF EXISTS idx_gba_scim_group_tenant;
DROP INDEX IF EXISTS idx_gba_scim_user_active;
DROP INDEX IF EXISTS idx_gba_scim_user_username;
DROP INDEX IF EXISTS idx_gba_scim_user_external;
DROP INDEX IF EXISTS idx_gba_scim_user_tenant;
DROP INDEX IF EXISTS idx_gba_scim_config_enabled;
DROP INDEX IF EXISTS idx_gba_scim_config_tenant;
DROP INDEX IF EXISTS idx_gba_saml_session_expires;
DROP INDEX IF EXISTS idx_gba_saml_session_tenant;
DROP INDEX IF EXISTS idx_gba_saml_session_user;
DROP INDEX IF EXISTS idx_gba_saml_config_enabled;
DROP INDEX IF EXISTS idx_gba_saml_config_tenant;

-- Drop tables
DROP TABLE IF EXISTS gba_scim_groups;
DROP TABLE IF EXISTS gba_scim_users;
DROP TABLE IF EXISTS gba_scim_configs;
DROP TABLE IF EXISTS gba_saml_sessions;
DROP TABLE IF EXISTS gba_saml_configs;
