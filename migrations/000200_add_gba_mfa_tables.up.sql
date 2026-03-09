-- GoBetterAuth MFA Tables
-- Creates tables for TOTP-based multi-factor authentication

-- TOTP configuration table
CREATE TABLE IF NOT EXISTS gba_mfa_totp (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES gba_users(id) ON DELETE CASCADE,
    secret TEXT NOT NULL,  -- Encrypted TOTP secret
    enabled BOOLEAN DEFAULT FALSE,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id)
);

-- Backup codes table for MFA recovery
CREATE TABLE IF NOT EXISTS gba_mfa_backup_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES gba_users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,  -- bcrypt hash of the backup code
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, code_hash)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_gba_mfa_totp_user ON gba_mfa_totp(user_id);
CREATE INDEX IF NOT EXISTS idx_gba_mfa_totp_enabled ON gba_mfa_totp(user_id, enabled) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_gba_mfa_totp_verified ON gba_mfa_totp(user_id, verified, enabled) WHERE verified = true AND enabled = true;

CREATE INDEX IF NOT EXISTS idx_gba_mfa_backup_user ON gba_mfa_backup_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_gba_mfa_backup_unused ON gba_mfa_backup_codes(user_id, used) WHERE used = false;

-- Add comment for documentation
COMMENT ON TABLE gba_mfa_totp IS 'TOTP MFA configuration for GoBetterAuth users';
COMMENT ON TABLE gba_mfa_backup_codes IS 'Single-use backup codes for MFA recovery';
COMMENT ON COLUMN gba_mfa_totp.secret IS 'Encrypted TOTP secret - decrypt before use';
COMMENT ON COLUMN gba_mfa_backup_codes.code_hash IS 'bcrypt hash of backup code - verify but never expose';
