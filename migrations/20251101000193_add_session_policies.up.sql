-- Add session policies to tenants table
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS session_max_duration INT DEFAULT 1440;  -- minutes (24 hours)
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS session_idle_timeout INT DEFAULT 480;   -- minutes (8 hours)
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS concurrent_sessions INT DEFAULT 5;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS session_persistence VARCHAR(20) DEFAULT 'device';  -- device, browser

-- Create indexes for session policy queries
CREATE INDEX IF NOT EXISTS idx_tenants_session_max_duration ON tenants(session_max_duration);
CREATE INDEX IF NOT EXISTS idx_tenants_session_idle_timeout ON tenants(session_idle_timeout);
