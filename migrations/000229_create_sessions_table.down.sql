-- Drop sessions table and related indexes
DROP INDEX IF EXISTS idx_sessions_user_active;
DROP INDEX IF EXISTS idx_sessions_last_activity;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_session_token;
DROP INDEX IF EXISTS idx_sessions_user_id;

DROP TABLE IF EXISTS sessions;