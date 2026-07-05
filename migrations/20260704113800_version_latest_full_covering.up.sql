-- Fully covering index for GetLatestVersions query.
-- Eliminates heap access entirely for the DISTINCT ON (function_id) query.
-- At ~1119 rows and ~934 bytes avg manifest, this adds ~1MB to index size — negligible.
-- Columns: function_id + published_at (index key for DISTINCT ON + ORDER BY) + all selected columns.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_registry_function_versions_latest_version_covering
ON registry_function_versions (function_id, published_at DESC)
INCLUDE (
    id,
    version,
    manifest,
    runtime,
    timeout_ms,
    memory_mb,
    deterministic,
    cache_ttl,
    capabilities,
    side_effects,
    idempotent,
    deployment_id,
    backend_id,
    content_hash,
    source_hash,
    bundle_size,
    updated_at
);
