-- Rollback: vault_security_hardening_phase1
BEGIN;

ALTER TABLE secrets_vault
    DROP COLUMN IF EXISTS kdf_params,
    DROP COLUMN IF EXISTS kdf_method,
    DROP COLUMN IF EXISTS expired_notified_at,
    DROP COLUMN IF EXISTS last_expiry_warning_at,
    DROP COLUMN IF EXISTS expire_after_days,
    DROP COLUMN IF EXISTS auto_expire,
    DROP COLUMN IF EXISTS status;

ALTER TABLE secret_access_tokens
    DROP COLUMN IF EXISTS ip_restriction_enabled,
    DROP COLUMN IF EXISTS denied_ips,
    DROP COLUMN IF EXISTS allowed_ips;

DROP TABLE IF EXISTS vault_escrow_config;
DROP TABLE IF EXISTS break_glass_config;
DROP TABLE IF EXISTS break_glass_requests;
DROP TABLE IF EXISTS vault_mfa_config;

COMMIT;
