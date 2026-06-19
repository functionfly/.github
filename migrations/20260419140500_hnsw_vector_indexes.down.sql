-- Migration: Rollback HNSW Vector Indexes
-- Reverts to ivfflat or removes vector indexes

-- ============================================
-- 1. Drop Functions
-- ============================================

DROP FUNCTION IF EXISTS hybrid_search_graph_definitions(TEXT, vector, INTEGER, FLOAT);
DROP FUNCTION IF EXISTS search_functions_by_code_similarity(vector, INTEGER, FLOAT);
DROP FUNCTION IF EXISTS search_team_memories_by_vector(UUID, vector, INTEGER, FLOAT);
DROP FUNCTION IF EXISTS search_agent_memories_by_vector(vector, INTEGER, FLOAT);
DROP FUNCTION IF EXISTS search_graph_definitions_by_vector(vector, INTEGER, FLOAT);

-- ============================================
-- 2. Drop Views
-- ============================================

DROP VIEW IF EXISTS hnsw_index_stats;

-- ============================================
-- 3. Drop HNSW Indexes
-- ============================================

DROP INDEX IF EXISTS idx_content_pages_embedding_hnsw;
DROP INDEX IF EXISTS idx_graph_exec_execution_embedding_hnsw;
DROP INDEX IF EXISTS idx_registry_functions_code_embedding_hnsw;
DROP INDEX IF EXISTS idx_team_memories_embedding_hnsw;
DROP INDEX IF EXISTS idx_agent_memories_embedding_hnsw;
DROP INDEX IF EXISTS idx_graph_definitions_embedding_hnsw;

-- ============================================
-- 4. Recreate ivfflat Indexes (Original Style)
-- ============================================

-- Recreate ivfflat index for graph definitions
CREATE INDEX IF NOT EXISTS idx_graph_definitions_ai_embedding 
ON graph_definitions 
USING ivfflat (ai_embedding vector_cosine_ops)
WITH (lists = 100);

COMMENT ON INDEX idx_graph_definitions_ai_embedding IS 
'IVFFlat index for graph definitions (reverted from HNSW). lists=100.';

-- ============================================
-- 5. Optional: Drop Embedding Columns (if needed)
-- ============================================

-- Uncomment to fully remove embedding support:
-- ALTER TABLE graph_execution_instances DROP COLUMN IF EXISTS execution_embedding;
-- ALTER TABLE registry_functions DROP COLUMN IF EXISTS code_embedding;
-- ALTER TABLE content_pages DROP COLUMN IF EXISTS content_embedding;

-- Keep GIN fallback index for agent_memories
CREATE INDEX IF NOT EXISTS idx_agent_memories_embedding_gin 
ON agent_memories USING GIN (embedding);
