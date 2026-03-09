-- Add MFA (Multi-Factor Authentication) support to users table
-- This migration adds the necessary columns for TOTP-based MFA

ALTER TABLE users
ADD COLUMN IF NOT EXISTS mfa_secret TEXT,
ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS mfa_backup_codes JSONB,
ADD COLUMN IF NOT EXISTS mfa_last_used TIMESTAMP WITH TIME ZONE;

-- Create index for MFA-enabled users for performance
CREATE INDEX IF NOT EXISTS idx_users_mfa_enabled ON users(mfa_enabled) WHERE mfa_enabled = true;

-- Create index for MFA last used for audit purposes
CREATE INDEX IF NOT EXISTS idx_users_mfa_last_used ON users(mfa_last_used) WHERE mfa_last_used IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN users.mfa_secret IS 'TOTP secret key for Multi-Factor Authentication';
COMMENT ON COLUMN users.mfa_enabled IS 'Whether MFA is enabled for this user';
COMMENT ON COLUMN users.mfa_backup_codes IS 'JSON array of hashed backup codes for MFA recovery';
COMMENT ON COLUMN users.mfa_last_used IS 'Timestamp of last MFA verification';