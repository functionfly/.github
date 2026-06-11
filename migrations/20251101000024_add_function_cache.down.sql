-- Rollback function cache table migration
-- Removes all cache data and table

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS idx_function_version_expires;
DROP INDEX IF EXISTS idx_last_hit_at;
DROP INDEX IF EXISTS idx_expires_at;
DROP INDEX IF EXISTS idx_function_version;
DROP INDEX IF EXISTS idx_cache_key;

-- Drop the table (this deletes all cached data)
DROP TABLE IF EXISTS function_cache;
