-- Rollback tenant_auth_settings table and related auth tables

DROP TABLE IF EXISTS tenant_auth_audit_log;
DROP TABLE IF EXISTS tenant_memberships;
DROP TABLE IF EXISTS tenant_invite_codes;
DROP TABLE IF EXISTS tenant_oauth_providers;
DROP TABLE IF EXISTS tenant_auth_settings;