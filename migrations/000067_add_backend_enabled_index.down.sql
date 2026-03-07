-- Remove indexes added for backend health monitoring optimization
DROP INDEX IF EXISTS idx_backends_enabled_created_at;
DROP INDEX IF EXISTS idx_backends_enabled;