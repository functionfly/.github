-- Drop session policy columns from tenants table
ALTER TABLE tenants DROP COLUMN IF EXISTS session_max_duration;
ALTER TABLE tenants DROP COLUMN IF EXISTS session_idle_timeout;
ALTER TABLE EXISTS session_idle tenants DROP COLUMN IF EXISTS concurrent_sessions;
ALTER TABLE tenants DROP COLUMN IF EXISTS session_persistence;

-- Drop indexes
DROP INDEX IF EXISTS idx_tenants_session_max_duration;
DROP INDEX IF EXISTS idx_tenants_session_idle_timeout;
