-- Add function cache table for L2 persistent caching
-- This enables caching of deterministic function execution results

CREATE TABLE IF NOT EXISTS function_cache (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Unique cache key (used for fast lookups)
    cache_key VARCHAR(255) NOT NULL UNIQUE,

    -- Cache key components (for targeted invalidation)
    function_id UUID NOT NULL,
    version VARCHAR(50) NOT NULL,
    input_hash VARCHAR(64) NOT NULL,

    -- Cached data
    output_json JSONB NOT NULL,
    output_size INTEGER NOT NULL,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    hit_count INTEGER NOT NULL DEFAULT 0,
    last_hit_at TIMESTAMPTZ,

    -- Constraint for unique function+version+input combination
    CONSTRAINT unique_function_version_input UNIQUE (function_id, version, input_hash)
);

-- Index for fast cache key lookups (primary lookup path)
CREATE INDEX IF NOT EXISTS idx_cache_key ON function_cache(cache_key);

-- Index for function+version lookups (invalidation queries)
CREATE INDEX IF NOT EXISTS idx_function_version ON function_cache(function_id, version);

-- Index for TTL-based cleanup queries
CREATE INDEX IF NOT EXISTS idx_expires_at ON function_cache(expires_at);

-- Index for hit count queries
CREATE INDEX IF NOT EXISTS idx_last_hit_at ON function_cache(last_hit_at);

-- Index for monitoring queries (function_id with expires)
CREATE INDEX IF NOT EXISTS idx_function_version_expires ON function_cache(function_id, version, expires_at);
