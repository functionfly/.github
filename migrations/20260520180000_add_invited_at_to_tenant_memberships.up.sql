-- Add missing invited_at and joined_at columns to tenant_memberships
-- The application code expects these columns but the current schema only has created_at/updated_at

ALTER TABLE tenant_memberships ADD COLUMN IF NOT EXISTS invited_at TIMESTAMPTZ;
ALTER TABLE tenant_memberships ADD COLUMN IF NOT EXISTS joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW();