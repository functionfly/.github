-- Migration: Create search execution tracking tables
-- Created: 20260608143000

-- Create search_executions table
CREATE TABLE IF NOT EXISTS search_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tool_name VARCHAR(50) NOT NULL,
    query TEXT NOT NULL,
    parameters JSONB,
    results_count INT DEFAULT 0,
    credits_used DECIMAL(10,4) NOT NULL DEFAULT 0,
    execution_time_ms INT NOT NULL DEFAULT 0,
    agent_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_search_executions_agent ON search_executions(agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_search_executions_tool ON search_executions(tool_name, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_search_executions_created ON search_executions(created_at DESC);

-- Add foreign key constraint if agent_identities table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'agent_identities') THEN
        ALTER TABLE search_executions
        ADD CONSTRAINT fk_search_executions_agent
        FOREIGN KEY (agent_id)
        REFERENCES agent_identities(id)
        ON DELETE SET NULL;
    END IF;
EXCEPTION WHEN undefined_table THEN
    -- Table doesn't exist yet, skip constraint
    NULL;
END $$;

-- Create search_result_cache table
CREATE TABLE IF NOT EXISTS search_result_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key VARCHAR(255) UNIQUE NOT NULL,
    tool_name VARCHAR(50) NOT NULL,
    query_hash VARCHAR(64) NOT NULL,
    parameters JSONB,
    results JSONB NOT NULL,
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Create indexes for cache management
CREATE INDEX IF NOT EXISTS idx_search_cache_key ON search_result_cache(cache_key);
CREATE INDEX IF NOT EXISTS idx_search_cache_expiry ON search_result_cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_search_cache_tool ON search_result_cache(tool_name);

-- Create function to auto-cleanup expired cache entries
CREATE OR REPLACE FUNCTION cleanup_expired_search_cache()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM search_result_cache
    WHERE expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create index for cleanup queries
CREATE INDEX IF NOT EXISTS idx_search_cache_expiry_asc ON search_result_cache(expires_at ASC) WHERE expires_at >= NOW();

COMMENT ON TABLE search_executions IS 'Tracks all search function executions for billing and analytics';
COMMENT ON TABLE search_result_cache IS 'Caches search results to reduce provider API calls and improve latency';
COMMENT ON FUNCTION cleanup_expired_search_cache IS 'Removes expired cache entries, returns count of deleted rows';