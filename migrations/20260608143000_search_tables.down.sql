-- Migration: Drop search tables (down migration)
-- Created: 20260608143000

DROP INDEX IF EXISTS idx_search_cache_expiry_asc;
DROP FUNCTION IF EXISTS cleanup_expired_search_cache();
DROP TABLE IF EXISTS search_result_cache;
DROP TABLE IF EXISTS search_executions;