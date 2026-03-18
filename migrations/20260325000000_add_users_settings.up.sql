-- Add settings JSONB to users for profile/notification preferences (visibility, notifications, privacy)
ALTER TABLE users ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{}';
COMMENT ON COLUMN users.settings IS 'User profile settings: visibility, notifications (deploymentSuccess, etc.), privacy';
