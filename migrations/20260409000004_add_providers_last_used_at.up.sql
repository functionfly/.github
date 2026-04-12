-- Migration: 20260409000004_add_providers_last_used_at
-- Description: Add last_used_at column to providers table for tracking stale connections
-- Problem: We can't track when a provider was last actively used
-- Solution: Add last_used_at timestamp column with index for efficient queries

BEGIN;

-- Add last_used_at column to providers table
ALTER TABLE providers
ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP WITH TIME ZONE NULL;

-- Create index for efficient queries on last_used_at
CREATE INDEX IF NOT EXISTS idx_providers_last_used_at
ON providers(last_used_at)
WHERE last_used_at IS NOT NULL;

-- Create index for finding stale providers (providers that haven't been used recently)
CREATE INDEX IF NOT EXISTS idx_providers_stale_check
ON providers(user_id, provider, last_used_at)
WHERE status = 'active' AND last_used_at IS NOT NULL;

COMMIT;
