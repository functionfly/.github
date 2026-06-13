-- Migration: add_slow_query_performance_indexes
-- Created at: 2026-06-11T21:21:00-05:00
-- Purpose: Add indexes to fix slow queries identified in performance analysis
--
-- Query 1: SELECT DISTINCT tenant_id FROM secrets_vault WHERE deleted_at IS NULL (710ms)
--   - Added partial index on (tenant_id) WHERE deleted_at IS NULL
--
-- Query 2: SELECT id FROM registry_function_executions WHERE timestamp < cutoff LIMIT 1000 (1017ms)
--   - Added covering index on (timestamp) INCLUDE (id) for index-only scan

-- Up migration (CONCURRENTLY requires autocommit, no transaction block)
-- Partial index for secrets_vault DISTINCT tenant_id query
-- This allows efficient retrieval of unique tenant_ids from non-deleted secrets
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_secrets_vault_active_tenant
ON secrets_vault(tenant_id)
WHERE deleted_at IS NULL;

-- Covering index for registry_function_executions retention cleanup queries
-- INCLUDE (id) enables index-only scan since the query only selects id
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_registry_function_executions_timestamp_id
ON registry_function_executions(timestamp)
INCLUDE (id);