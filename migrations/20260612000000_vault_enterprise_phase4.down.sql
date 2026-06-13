-- Rollback: vault_enterprise_phase4
BEGIN;

DROP TABLE IF EXISTS vault_siem_webhooks;
DROP TABLE IF EXISTS vault_sso_config;
DROP TABLE IF EXISTS vault_shares;
DROP TABLE IF EXISTS vault_role_assignments;
DROP TABLE IF EXISTS vault_roles;
DROP TABLE IF EXISTS vault_namespaces;

ALTER TABLE secrets_vault
    DROP COLUMN IF EXISTS is_shared,
    DROP COLUMN IF EXISTS namespace;

COMMIT;
