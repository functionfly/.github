-- Migration: Revert HNSW index optimizations for team_memories
-- Reverses: 20260411191500_team_memory_hnsw_optimization

-- Drop the optimized indexes
DROP INDEX IF EXISTS idx_team_memories_embedding_hnsw;
DROP INDEX IF EXISTS idx_team_memories_embedding_encrypted_hnsw;
DROP INDEX IF EXISTS idx_team_memories_team_type_embedding;
DROP INDEX IF EXISTS idx_similarity_clusters_memory_id;

-- Drop the materialized view
DROP MATERIALIZED VIEW IF EXISTS team_memory_similarity_clusters;

-- Drop helper functions
DROP FUNCTION IF EXISTS set_hnsw_ef_search(integer);
DROP FUNCTION IF EXISTS search_team_memories_adaptive(vector, integer, float);

-- Recreate original HNSW index with default parameters
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
    CREATE INDEX idx_team_memories_embedding_hnsw
    ON team_memories USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
  END IF;
EXCEPTION
  WHEN OTHERS THEN NULL;
END
$$;