-- Reverse vault security hardening

-- Remove cleanup function
DROP FUNCTION IF EXISTS cleanup_expired_vault_tokens;

-- Remove audit immutability trigger
DROP TRIGGER IF EXISTS trg_audit_log_immutable ON secrets_audit_log;
DROP FUNCTION IF EXISTS prevent_audit_log_modification();

-- Note: Column type conversions (VARCHAR→BYTEA) are not easily reversible
-- and are safe to leave in place. The data format is compatible in both directions.
