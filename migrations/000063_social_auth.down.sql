-- Remove social authentication fields from users table
DROP INDEX IF EXISTS idx_users_provider_provider_id;
DROP INDEX IF EXISTS idx_users_provider;

ALTER TABLE users DROP COLUMN IF EXISTS provider_data;
ALTER TABLE users DROP COLUMN IF EXISTS provider_id;
ALTER TABLE users DROP COLUMN IF EXISTS provider;