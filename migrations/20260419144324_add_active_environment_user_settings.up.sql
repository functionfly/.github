-- Add active environment support to user settings
-- Stores the user's currently selected environment (production, staging, development)

-- Add GIN index on settings JSONB for efficient querying of active_environment
CREATE INDEX IF NOT EXISTS idx_users_settings_active_env 
ON users USING GIN ((settings -> 'active_environment'));

COMMENT ON INDEX idx_users_settings_active_env IS 'Index for querying users by their selected active environment';

-- Update existing users to have production as default if not set
UPDATE users 
SET settings = jsonb_set(
    COALESCE(settings, '{}'::jsonb),
    '{active_environment}',
    '"production"'::jsonb,
    true
)
WHERE settings IS NULL OR settings -> 'active_environment' IS NULL;

-- Add index for tenant-level environment analytics
CREATE INDEX IF NOT EXISTS idx_users_tenant_id_settings_env 
ON users(tenant_id, (settings ->> 'active_environment')) 
WHERE settings ->> 'active_environment' IS NOT NULL;
