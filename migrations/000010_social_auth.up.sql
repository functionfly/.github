-- Add social authentication fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_data JSONB;

-- Add unique constraint to prevent duplicate social accounts
-- Allow NULL values but ensure uniqueness when not NULL
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider_provider_id ON users(provider, provider_id) WHERE provider IS NOT NULL AND provider_id IS NOT NULL;

-- Add index for efficient lookups by provider
CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider) WHERE provider IS NOT NULL;