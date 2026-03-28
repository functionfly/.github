-- Token version column for enhanced session security
-- Enables server-side token invalidation on password change or security events

ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version INTEGER DEFAULT 0 NOT NULL;
ALTER TABLE users ALTER COLUMN token_version DROP DEFAULT;

CREATE INDEX IF NOT EXISTS idx_users_token_version ON users(token_version);
