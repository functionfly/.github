-- Migration: Drop admin_sessions table
-- This reverts the creation of the admin_sessions table

DROP INDEX IF EXISTS idx_admin_sessions_last_activity;
DROP INDEX IF EXISTS idx_admin_sessions_expires_at;
DROP INDEX IF EXISTS idx_admin_sessions_token_hash;
DROP INDEX IF EXISTS idx_admin_sessions_user_id;
DROP TABLE IF EXISTS admin_sessions;
