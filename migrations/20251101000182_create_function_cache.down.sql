-- Migration: Drop function_cache table
-- Rolls back the execution caching L2 disk cache

DROP INDEX IF EXISTS idx_function_cache_last_hit_at;
DROP INDEX IF EXISTS idx_function_cache_expires_at;
DROP INDEX IF EXISTS idx_function_cache_function_version;
DROP TABLE IF EXISTS function_cache;
