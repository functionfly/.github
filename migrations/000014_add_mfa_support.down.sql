-- Remove MFA (Multi-Factor Authentication) support from users table

-- Drop indexes first
DROP INDEX IF EXISTS idx_users_mfa_last_used;
DROP INDEX IF EXISTS idx_users_mfa_enabled;

-- Remove MFA columns
ALTER TABLE users
DROP COLUMN IF EXISTS mfa_last_used,
DROP COLUMN IF EXISTS mfa_backup_codes,
DROP COLUMN IF EXISTS mfa_enabled,
DROP COLUMN IF EXISTS mfa_secret;