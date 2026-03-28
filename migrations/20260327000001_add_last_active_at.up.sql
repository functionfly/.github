-- Add last_active_at column to users table for online status tracking
-- This column stores the timestamp of when the user was last active

ALTER TABLE users ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMP WITH TIME ZONE;

-- Create index for efficient online status queries
CREATE INDEX IF NOT EXISTS idx_users_last_active_at ON users(last_active_at);

-- Create a GIN index for querying online users (active in last 5 minutes)
-- Note: This is a partial index that only includes users active in the last day
CREATE INDEX IF NOT EXISTS idx_users_recently_active ON users(last_active_at)
    WHERE last_active_at > NOW() - INTERVAL '1 day';
