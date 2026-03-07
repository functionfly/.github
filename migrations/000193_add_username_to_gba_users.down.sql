-- Rollback username column from gba_users table

DROP INDEX IF EXISTS idx_users_username_tenant;
ALTER TABLE gba_users DROP COLUMN IF EXISTS username;
