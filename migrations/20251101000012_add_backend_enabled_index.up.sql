-- Add index on backends.enabled to optimize health monitoring queries
-- This prevents full table scans when fetching enabled backends for health checks
CREATE INDEX IF NOT EXISTS idx_backends_enabled ON backends(enabled);

-- Add composite index on (enabled, created_at) for optimal ordering performance
CREATE INDEX IF NOT EXISTS idx_backends_enabled_created_at ON backends(enabled, created_at);