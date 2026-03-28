-- Remove last_active_at column from users table

DROP INDEX IF EXISTS idx_users_recently_active;
DROP INDEX IF EXISTS idx_users_last_active_at;
ALTER TABLE users DROP COLUMN IF EXISTS last_active_at;
