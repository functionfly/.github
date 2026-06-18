-- Rollback for vault client-encrypted dynamic credentials.
-- Note: dropping columns can fail if they contain data. This is the
-- canonical "down" for symmetry; do not run in production without
-- confirming downstream code can tolerate the removal.

DROP TABLE IF EXISTS dynamic_target_shares;
DROP TABLE IF EXISTS dynamic_wrapped_access_tokens;
DROP TABLE IF EXISTS vault_user_keys;
DROP TABLE IF EXISTS vault_tenant_keys;

ALTER TABLE dynamic_secret_targets
  DROP CONSTRAINT IF EXISTS chk_dynamic_secret_targets_encryption_mode;
ALTER TABLE dynamic_secret_targets
  DROP COLUMN IF EXISTS encryption_mode,
  DROP COLUMN IF EXISTS key_version,
  DROP COLUMN IF EXISTS wrap_iv,
  DROP COLUMN IF EXISTS wrap_auth_tag,
  DROP COLUMN IF EXISTS namespace;
