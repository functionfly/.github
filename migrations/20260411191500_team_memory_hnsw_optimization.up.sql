-- Migration: Optimize HNSW index parameters for team_memories vector search at scale
-- Created: 2026-04-11
-- Purpose: Tune HNSW parameters for higher recall and faster search with large datasets

-- HNSW Parameter Selection Rationale:
-- m = 32: Doubles the default (16) for better connectivity at scale (10K+ vectors)
-- ef_construction = 128: Higher quality index build (vs 64 default) for better recall
-- ef = 64: Search-time exploration factor (queries will use SET hnsw.ef = 64)

-- Recreate the HNSW index with optimized parameters for production scale
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
    -- Drop existing index if it exists (safe to recreate with better params)
    DROP INDEX IF EXISTS idx_team_memories_embedding_hnsw;
    
    -- Create optimized HNSW index for cosine similarity search
    -- m=32: Each element connected to 32 neighbors (better for 10K+ vectors)
    -- ef_construction=128: Higher quality during build (slower build, faster/better search)
    CREATE INDEX idx_team_memories_embedding_hnsw
    ON team_memories USING hnsw (embedding vector_cosine_ops)
    WITH (m = 32, ef_construction = 128);
    
    -- Also create a smaller, faster index for encrypted memories (embedding-only search)
    -- These have no content to search, so we want pure vector similarity
    DROP INDEX IF EXISTS idx_team_memories_embedding_encrypted_hnsw;
    
    CREATE INDEX idx_team_memories_embedding_encrypted_hnsw
    ON team_memories USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64)
    WHERE is_encrypted = true;
  END IF;
EXCEPTION
  WHEN OTHERS THEN 
    RAISE NOTICE 'Failed to create HNSW indexes: %', SQLERRM;
END
$$;

-- Add helper function for setting HNSW search parameters per query
CREATE OR REPLACE FUNCTION set_hnsw_ef_search(ef integer)
RETURNS void AS $$
BEGIN
  -- Set the ef search parameter for the current session
  -- Higher ef = better recall, slower search
  -- Recommended: ef = 64 for balanced, ef = 100+ for high recall
  SET LOCAL hnsw.ef_search = ef;
END;
$$ LANGUAGE plpgsql;

-- Add function for adaptive HNSW search based on result size needed
CREATE OR REPLACE FUNCTION search_team_memories_adaptive(
  query_vector vector(1536),
  result_limit integer,
  min_relevance float DEFAULT 0.7
)
RETURNS TABLE (
  id UUID,
  team_id UUID,
  memory_type VARCHAR(50),
  summary TEXT,
  relevance_score FLOAT
) AS $$
DECLARE
  ef_setting integer;
BEGIN
  -- Adaptive ef based on result limit (ef should be >= limit)
  ef_setting := GREATEST(result_limit * 2, 64);
  ef_setting := LEAST(ef_setting, 200); -- Cap at 200
  
  -- Set search parameter
  PERFORM set_config('hnsw.ef_search', ef_setting::text, true);
  
  -- Return results with relevance score
  RETURN QUERY
  SELECT 
    m.id,
    m.team_id,
    m.memory_type,
    m.summary,
    (1.0 - (m.embedding <=> query_vector))::float as relevance_score
  FROM team_memories m
  WHERE m.is_encrypted = false
    AND (m.expires_at IS NULL OR m.expires_at > NOW())
    AND (1.0 - (m.embedding <=> query_vector)) >= min_relevance
  ORDER BY m.embedding <=> query_vector
  LIMIT result_limit;
END;
$$ LANGUAGE plpgsql;

-- Create a materialized view for pre-computed vector similarity clusters
-- This enables fast "related memories" lookups without HNSW overhead
CREATE MATERIALIZED VIEW IF NOT EXISTS team_memory_similarity_clusters AS
WITH memory_pairs AS (
  SELECT 
    m1.id as memory_id,
    m1.team_id,
    m1.memory_type,
    ARRAY_AGG(
      m2.id ORDER BY m1.embedding <=> m2.embedding
    ) FILTER (WHERE m1.id != m2.id AND (1.0 - (m1.embedding <=> m2.embedding)) > 0.85) as related_ids,
    ARRAY_AGG(
      (1.0 - (m1.embedding <=> m2.embedding))::float 
      ORDER BY m1.embedding <=> m2.embedding
    ) FILTER (WHERE m1.id != m2.id AND (1.0 - (m1.embedding <=> m2.embedding)) > 0.85) as related_scores
  FROM team_memories m1
  JOIN team_memories m2 ON m1.team_id = m2.team_id 
    AND m1.is_encrypted = false 
    AND m2.is_encrypted = false
  WHERE m1.is_encrypted = false
  GROUP BY m1.id, m1.team_id, m1.memory_type
)
SELECT * FROM memory_pairs
WHERE array_length(related_ids, 1) > 0;

-- Index for the materialized view
CREATE INDEX IF NOT EXISTS idx_similarity_clusters_memory_id 
ON team_memory_similarity_clusters(memory_id);

-- Create index for team+type filtered vector searches (common query pattern)
CREATE INDEX IF NOT EXISTS idx_team_memories_team_type_embedding 
ON team_memories USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64)
WHERE is_encrypted = false AND memory_type IN ('decision', 'preference');

-- Comments explaining the optimization strategy
COMMENT ON INDEX idx_team_memories_embedding_hnsw IS 
  'Optimized HNSW index: m=32, ef_construction=128 for high-quality ANN search at 10K+ vectors scale. Build is slower but recall is significantly improved.';

COMMENT ON INDEX idx_team_memories_embedding_encrypted_hnsw IS 
  'Lightweight HNSW for encrypted memories: m=16, ef_construction=64. These have smaller query volume (no text search fallback).';

COMMENT ON FUNCTION search_team_memories_adaptive IS 
  'Adaptive HNSW search with dynamic ef parameter based on result limit. Automatically adjusts exploration factor for optimal speed/recall balance.';
