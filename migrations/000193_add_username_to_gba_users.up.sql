-- Add username column to gba_users table for GoBetterAuth
-- Allows usernames with tenant-scoped uniqueness for real-time availability checking

ALTER TABLE gba_users ADD COLUMN IF NOT EXISTS username VARCHAR(255);

-- Create unique index for username + tenant_id combination
-- This ensures usernames are unique within a tenant, not globally
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_tenant
ON gba_users(tenant_id, username)
WHERE username IS NOT NULL;
