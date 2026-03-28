DROP INDEX IF EXISTS idx_users_token_version;
ALTER TABLE users DROP COLUMN IF EXISTS token_version;
