-- Rollback: dynamic_secrets_phase2
BEGIN;
DROP TABLE IF EXISTS dynamic_credential_leases;
DROP TABLE IF EXISTS dynamic_credentials;
DROP TABLE IF EXISTS dynamic_secret_targets;
COMMIT;
