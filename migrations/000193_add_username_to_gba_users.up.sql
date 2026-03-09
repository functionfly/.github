-- Add username column to gba_users table for GoBetterAuth
-- Allows usernames with tenant-scoped uniqueness for real-time availability checking
-- Only runs when gba_users exists (add_gba_base_tables may run after this in version order)

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'gba_users') THEN
    ALTER TABLE gba_users ADD COLUMN IF NOT EXISTS username VARCHAR(255);
    CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_tenant ON gba_users(tenant_id, username) WHERE username IS NOT NULL;
  END IF;
END $$;
