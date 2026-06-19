-- Add MFA enforcement columns to tenants and users
-- This migration adds support for tenant-level and user-level MFA policy enforcement

-- Add MFA policy column to tenants (optional, required, suspended)
ALTER TABLE tenants ADD COLUMN mfa_policy VARCHAR(20) DEFAULT 'optional';

-- Add MFA enforcement flag to users (for user-level enforcement)
ALTER TABLE users ADD COLUMN mfa_enforced BOOLEAN DEFAULT false;

-- Create index for faster MFA policy lookups
CREATE INDEX IF NOT EXISTS idx_tenants_mfa_policy ON tenants(mfa_policy) WHERE mfa_policy IS NOT NULL;

-- Create index for faster user MFA enforcement lookups
CREATE INDEX IF NOT EXISTS idx_users_mfa_enforced ON users(mfa_enforced) WHERE mfa_enforced = true;
