-- Add missing columns to tenant_memberships and users tables
-- These were expected by the application code but not present in the current schema

-- tenant_memberships: last_active_at for tracking member activity
ALTER TABLE tenant_memberships ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMPTZ;

-- users: settings JSONB for user preferences (used by /v1/users/me/settings endpoint)
ALTER TABLE users ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}';
