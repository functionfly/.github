-- Rollback GoBetterAuth MFA Tables
-- Removes tables created for TOTP-based multi-factor authentication

-- Drop indexes first
DROP INDEX IF EXISTS idx_gba_mfa_backup_unused;
DROP INDEX IF EXISTS idx_gba_mfa_backup_user;
DROP INDEX IF EXISTS idx_gba_mfa_totp_verified;
DROP INDEX IF EXISTS idx_gba_mfa_totp_enabled;
DROP INDEX IF EXISTS idx_gba_mfa_totp_user;

-- Drop tables
DROP TABLE IF EXISTS gba_mfa_backup_codes;
DROP TABLE IF EXISTS gba_mfa_totp;
