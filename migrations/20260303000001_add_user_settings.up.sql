-- Migration: Add user settings column
-- Created: 2026-03-03

-- Add settings JSONB column to users table for storing profile preferences
ALTER TABLE users ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{}';

-- Create index for efficient settings queries
CREATE INDEX IF NOT EXISTS idx_users_settings ON users USING GIN (settings);

-- Add comment for documentation
COMMENT ON COLUMN users.settings IS 'User profile settings including visibility, notifications, and privacy preferences';
