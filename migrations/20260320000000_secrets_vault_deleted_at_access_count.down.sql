ALTER TABLE secrets_vault
  DROP COLUMN IF EXISTS deleted_at,
  DROP COLUMN IF EXISTS access_count;
