DROP INDEX IF EXISTS idx_users_verification_code;
ALTER TABLE users DROP COLUMN IF EXISTS verification_code;
