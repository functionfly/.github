-- Migration: Create function_cache table for execution caching system
-- This provides L2 persistent caching for deterministic function outputs

CREATE TABLE IF NOT EXISTS function_cache (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key   VARCHAR(255) NOT NULL,
    function_id UUID NOT NULL,
    version     VARCHAR(50) NOT NULL,
    input_hash  VARCHAR(64) NOT NULL,
    output_json JSONB NOT NULL,
    output_size INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    hit_count   INTEGER NOT NULL DEFAULT 0,
    last_hit_at TIMESTAMPTZ,
    CONSTRAINT uq_function_cache_key UNIQUE (cache_key)
);

-- Indexes for efficient lookups and invalidation
CREATE INDEX IF NOT EXISTS idx_function_cache_function_version ON function_cache (function_id, version);
CREATE INDEX IF NOT EXISTS idx_function_cache_expires_at ON function_cache (expires_at);
CREATE INDEX IF NOT EXISTS idx_function_cache_last_hit_at ON function_cache (last_hit_at);

-- Add comment for documentation
COMMENT ON TABLE function_cache IS 'L2 disk cache for deterministic function execution results';
COMMENT ON COLUMN function_cache.cache_key IS 'Composite key: fx:cache:{function_id}:{version}:{input_hash}';
COMMENT ON COLUMN function_cache.input_hash IS 'SHA-256 hash of normalized JSON input';
