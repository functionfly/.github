-- Add deleted_at and access_count to secrets_vault for soft delete and usage tracking.
-- The Go vault repository expects these columns (GetSecretsByTenant filters on deleted_at IS NULL).

ALTER TABLE secrets_vault
  ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS access_count INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN secrets_vault.deleted_at IS 'Soft delete timestamp; NULL means active';
COMMENT ON COLUMN secrets_vault.access_count IS 'Number of times the secret was accessed';
