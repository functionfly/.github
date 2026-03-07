e-- Rollback MFA enforcement columns
-- This migration removes the MFA policy columns added in the up migration

-- Drop indexes first
DROP INDEX IF EXISTS idx_tenants_mfa_policy;
DROP INDEX IF EXISTS idx_users_mfa_enforced;

-- Drop MFA enforcement column from users
ALTER TABLE users DROP COLUMN IF EXISTS mfa_enforced;

-- Drop MFA policy column from tenants
ALTER TABLE tenants DROP COLUMN IF EXISTS mfa_policy;
