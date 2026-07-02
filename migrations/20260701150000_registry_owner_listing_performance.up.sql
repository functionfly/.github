-- Fix slow owner listing queries (~280ms → sub-ms)
-- Problem: idx_registry_functions_owner_user only has (owner_user_id),
-- so ORDER BY created_at DESC requires a sort over all matching rows.
-- Solution: composite index that covers both WHERE and ORDER BY,
-- enabling an index scan that returns rows already in the correct order.
-- With LIMIT 20, PostgreSQL can stop after 20 index entries.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_registry_functions_owner_created
ON registry_functions (owner_user_id, created_at DESC)
WHERE owner_user_id IS NOT NULL;

-- Fix slow DISTINCT ON latest version queries (~275ms → sub-ms)
-- Problem: idx_registry_function_versions_latest covers (function_id, published_at DESC)
-- but every selected column requires a random heap fetch (manifest is JSONB, large).
-- Solution: covering index that includes lightweight scalar columns,
-- eliminating heap access for the most common registry list queries.
-- Manifest/capabilities are intentionally excluded (too wide for INCLUDE).

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_registry_function_versions_latest_covering
ON registry_function_versions (function_id, published_at DESC)
INCLUDE (id, version, runtime, timeout_ms, memory_mb, deterministic, cache_ttl, side_effects, idempotent, bundle_size, is_active);
