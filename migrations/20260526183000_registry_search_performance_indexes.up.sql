-- Speed up batch latest-version lookups for registry search/gallery (100 IDs per page).
-- Index-only friendly columns for list cards; large blobs stay on heap when needed.
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_latest_listing
ON registry_function_versions (function_id, published_at DESC)
INCLUDE (id, version, runtime, timeout_ms, memory_mb, deterministic, cache_ttl, side_effects, idempotent, bundle_size, updated_at);

-- Join path used by GetGalleryStats and ListLatestVersionsForFunctions fast path.
CREATE INDEX IF NOT EXISTS idx_registry_functions_public_latest_version
ON registry_functions (id, latest_version)
WHERE visibility = 'public' AND latest_version IS NOT NULL AND latest_version <> '';
