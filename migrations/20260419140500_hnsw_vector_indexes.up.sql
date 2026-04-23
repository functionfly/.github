-- Migration: HNSW Vector Index Optimization
-- Upgrades pgvector indexes from ivfflat to HNSW for better recall and performance
-- Created: 2026-04-19

-- ============================================
-- Pre-requisite: Ensure pgvector extension is available
-- ============================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
        RAISE EXCEPTION 'pgvector extension not found. Please install pgvector first.';
    END IF;
END $$;

-- ============================================
-- 1. Graph Definitions - AI Embeddings (384-dim)
-- Semantic search for graph workflows
-- ============================================

-- Drop old ivfflat index if exists
DROP INDEX IF EXISTS idx_graph_definitions_ai_embedding;
DROP INDEX IF EXISTS idx_graph_definitions_embedding_ivfflat;

-- Create HNSW index for graph embeddings
-- m=16, ef_construction=64 for good build performance
-- ef=16 for search (can be increased at query time with SET hnsw.ef_search)
CREATE INDEX idx_graph_definitions_embedding_hnsw 
ON graph_definitions 
USING hnsw (ai_embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

COMMENT ON INDEX idx_graph_definitions_embedding_hnsw IS 
'HNSW index for semantic search on graph definitions. m=16, ef_construction=64. Better recall than ivfflat.';

-- ============================================
-- 2. Agent Memories - Structured Data Embeddings
-- AI agent memory with vector search
-- ============================================

-- Drop old GIN index on embedding arrays (fallback without pgvector)
DROP INDEX IF EXISTS idx_agent_memories_embedding_gin;

-- Create HNSW index for agent memory embeddings
-- Assuming 384-dim embeddings (same as graph definitions)
-- If different dimension, adjust accordingly
CREATE INDEX idx_agent_memories_embedding_hnsw 
ON agent_memories 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

COMMENT ON INDEX idx_agent_memories_embedding_hnsw IS 
'HNSW index for AI agent memory similarity search. Enables fast recall of relevant memories.';

-- ============================================
-- 3. Team Memories - Semantic Search
-- Cross-team knowledge sharing with vector search
-- ============================================

-- Create HNSW index for team memory embeddings
CREATE INDEX idx_team_memories_embedding_hnsw 
ON team_memories 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

COMMENT ON INDEX idx_team_memories_embedding_hnsw IS 
'HNSW index for team memory semantic search. Enables finding similar memories across teams.';

-- ============================================
-- 4. Function Registry - Code Embeddings
-- Semantic code search for functions
-- ============================================

-- Add embedding column to registry_functions if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'registry_functions' 
        AND column_name = 'code_embedding'
    ) THEN
        ALTER TABLE registry_functions 
        ADD COLUMN code_embedding vector(384);
        
        COMMENT ON COLUMN registry_functions.code_embedding IS 
        'Vector embedding of function code for semantic search. 384-dim.';
    END IF;
END $$;

-- Create HNSW index for code embeddings
CREATE INDEX idx_registry_functions_code_embedding_hnsw 
ON registry_functions 
USING hnsw (code_embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

COMMENT ON INDEX idx_registry_functions_code_embedding_hnsw IS 
'HNSW index for semantic code search. Enables finding similar functions by code semantics.';

-- ============================================
-- 5. Graph Execution Traces - Similarity Search
-- Find similar execution patterns
-- ============================================

-- Add embedding column for execution traces if useful
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'graph_execution_instances' 
        AND column_name = 'execution_embedding'
    ) THEN
        ALTER TABLE graph_execution_instances 
        ADD COLUMN execution_embedding vector(128);
        
        COMMENT ON COLUMN graph_execution_instances.execution_embedding IS 
        'Vector embedding of execution pattern for anomaly detection. 128-dim.';
    END IF;
END $$;

-- Create HNSW index for execution pattern embeddings
CREATE INDEX idx_graph_exec_execution_embedding_hnsw 
ON graph_execution_instances 
USING hnsw (execution_embedding vector_cosine_ops)
WITH (m = 12, ef_construction = 40);

COMMENT ON INDEX idx_graph_exec_execution_embedding_hnsw IS 
'HNSW index for execution pattern similarity. Lower dim requires smaller m.';

-- ============================================
-- 6. Content/Documentation - Semantic Search
-- Help docs, error messages, tutorials
-- ============================================

-- Check if content repository table exists and add embedding
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'content_pages'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'content_pages' 
        AND column_name = 'content_embedding'
    ) THEN
        ALTER TABLE content_pages 
        ADD COLUMN content_embedding vector(384);
        
        CREATE INDEX idx_content_pages_embedding_hnsw 
        ON content_pages 
        USING hnsw (content_embedding vector_cosine_ops)
        WITH (m = 16, ef_construction = 64);
        
        COMMENT ON INDEX idx_content_pages_embedding_hnsw IS 
        'HNSW index for documentation semantic search.';
    END IF;
END $$;

-- ============================================
-- 7. Query Functions for Vector Search
-- ============================================

-- Function to search graph definitions by embedding similarity
CREATE OR REPLACE FUNCTION search_graph_definitions_by_vector(
    p_query_embedding vector(384),
    p_limit INTEGER DEFAULT 10,
    p_min_similarity FLOAT DEFAULT 0.7
)
RETURNS TABLE (
    graph_id UUID,
    graph_name TEXT,
    similarity_score FLOAT
) AS $$
BEGIN
    -- Set ef_search for better recall at query time
    SET LOCAL hnsw.ef_search = 100;
    
    RETURN QUERY
    SELECT 
        gd.id as graph_id,
        gd.name::TEXT as graph_name,
        1 - (gd.ai_embedding <=> p_query_embedding)::FLOAT as similarity_score
    FROM graph_definitions gd
    WHERE gd.ai_embedding IS NOT NULL
        AND 1 - (gd.ai_embedding <=> p_query_embedding) > p_min_similarity
    ORDER BY gd.ai_embedding <=> p_query_embedding
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION search_graph_definitions_by_vector(vector, INTEGER, FLOAT) IS 
'Semantic search for graph definitions. Returns results above similarity threshold. Uses HNSW index.';

-- Function to search agent memories by vector
CREATE OR REPLACE FUNCTION search_agent_memories_by_vector(
    p_query_embedding vector(384),
    p_limit INTEGER DEFAULT 10,
    p_min_similarity FLOAT DEFAULT 0.7
)
RETURNS TABLE (
    memory_id UUID,
    memory_title TEXT,
    memory_type TEXT,
    agent_id TEXT,
    similarity_score FLOAT
) AS $$
BEGIN
    SET LOCAL hnsw.ef_search = 100;
    
    RETURN QUERY
    SELECT 
        am.id as memory_id,
        am.title::TEXT as memory_title,
        am.memory_type::TEXT,
        am.agent_id::TEXT,
        1 - (am.embedding <=> p_query_embedding)::FLOAT as similarity_score
    FROM agent_memories am
    WHERE am.embedding IS NOT NULL
        AND 1 - (am.embedding <=> p_query_embedding) > p_min_similarity
    ORDER BY am.embedding <=> p_query_embedding
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION search_agent_memories_by_vector(vector, INTEGER, FLOAT) IS 
'Semantic search for agent memories. Enables AI agents to recall relevant past experiences.';

-- Function to search team memories by vector
CREATE OR REPLACE FUNCTION search_team_memories_by_vector(
    p_team_id UUID,
    p_query_embedding vector(384),
    p_limit INTEGER DEFAULT 10,
    p_min_similarity FLOAT DEFAULT 0.7
)
RETURNS TABLE (
    memory_id UUID,
    title TEXT,
    content_preview TEXT,
    memory_type TEXT,
    similarity_score FLOAT
) AS $$
BEGIN
    SET LOCAL hnsw.ef_search = 100;
    
    RETURN QUERY
    SELECT 
        tm.id as memory_id,
        tm.title::TEXT,
        LEFT(tm.content, 200)::TEXT as content_preview,
        tm.memory_type::TEXT,
        1 - (tm.embedding <=> p_query_embedding)::FLOAT as similarity_score
    FROM team_memories tm
    WHERE tm.team_id = p_team_id
        AND tm.embedding IS NOT NULL
        AND 1 - (tm.embedding <=> p_query_embedding) > p_min_similarity
    ORDER BY tm.embedding <=> p_query_embedding
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION search_team_memories_by_vector(UUID, vector, INTEGER, FLOAT) IS 
'Semantic search within a team memory. Enables finding relevant team knowledge.';

-- Function to search registry functions by code similarity
CREATE OR REPLACE FUNCTION search_functions_by_code_similarity(
    p_code_embedding vector(384),
    p_limit INTEGER DEFAULT 10,
    p_min_similarity FLOAT DEFAULT 0.8
)
RETURNS TABLE (
    function_id UUID,
    function_name TEXT,
    author_username TEXT,
    similarity_score FLOAT
) AS $$
BEGIN
    SET LOCAL hnsw.ef_search = 100;
    
    RETURN QUERY
    SELECT 
        rf.id as function_id,
        rf.name::TEXT as function_name,
        rf.author_username::TEXT,
        1 - (rf.code_embedding <=> p_code_embedding)::FLOAT as similarity_score
    FROM registry_functions rf
    WHERE rf.code_embedding IS NOT NULL
        AND 1 - (rf.code_embedding <=> p_code_embedding) > p_min_similarity
    ORDER BY rf.code_embedding <=> p_code_embedding
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION search_functions_by_code_similarity(vector, INTEGER, FLOAT) IS 
'Semantic code search for finding similar functions. Helps developers discover existing solutions.';

-- ============================================
-- 8. Hybrid Search (Vector + Full-Text)
-- ============================================

-- Function for hybrid search combining vector similarity and text search
CREATE OR REPLACE FUNCTION hybrid_search_graph_definitions(
    p_query_text TEXT,
    p_query_embedding vector(384),
    p_limit INTEGER DEFAULT 10,
    p_vector_weight FLOAT DEFAULT 0.7
)
RETURNS TABLE (
    graph_id UUID,
    graph_name TEXT,
    description TEXT,
    vector_score FLOAT,
    text_score FLOAT,
    combined_score FLOAT
) AS $$
BEGIN
    SET LOCAL hnsw.ef_search = 100;
    
    RETURN QUERY
    WITH vector_search AS (
        SELECT 
            gd.id,
            1 - (gd.ai_embedding <=> p_query_embedding)::FLOAT as sim_score
        FROM graph_definitions gd
        WHERE gd.ai_embedding IS NOT NULL
        ORDER BY gd.ai_embedding <=> p_query_embedding
        LIMIT 100
    ),
    text_search AS (
        SELECT 
            gd.id,
            ts_rank(
                to_tsvector('english', COALESCE(gd.name, '') || ' ' || COALESCE(gd.description, '')),
                plainto_tsquery('english', p_query_text)
            ) as rank_score
        FROM graph_definitions gd
        WHERE to_tsvector('english', COALESCE(gd.name, '') || ' ' || COALESCE(gd.description, ''))
            @@ plainto_tsquery('english', p_query_text)
        ORDER BY rank_score DESC
        LIMIT 100
    )
    SELECT 
        gd.id as graph_id,
        gd.name::TEXT as graph_name,
        gd.description::TEXT,
        COALESCE(vs.sim_score, 0) as vector_score,
        COALESCE(ts.rank_score, 0) as text_score,
        (COALESCE(vs.sim_score, 0) * p_vector_weight + 
         COALESCE(ts.rank_score, 0) * (1 - p_vector_weight)) as combined_score
    FROM graph_definitions gd
    LEFT JOIN vector_search vs ON gd.id = vs.id
    LEFT JOIN text_search ts ON gd.id = ts.id
    WHERE vs.id IS NOT NULL OR ts.id IS NOT NULL
    ORDER BY combined_score DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION hybrid_search_graph_definitions(TEXT, vector, INTEGER, FLOAT) IS 
'Hybrid search combining vector similarity and full-text search. Weight adjustable (default 70% vector, 30% text).';

-- ============================================
-- 9. Vector Index Performance Monitoring
-- ============================================

-- View to track HNSW index stats
CREATE OR REPLACE VIEW hnsw_index_stats AS
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    pg_relation_size(indexrelid) as index_size_bytes,
    (SELECT count(*) FROM pg_stat_user_indexes WHERE indexrelname LIKE '%hnsw%') as total_hnsw_indexes
FROM pg_indexes
JOIN pg_class ON pg_class.relname = indexname
WHERE indexname LIKE '%hnsw%'
AND schemaname = 'public';

COMMENT ON VIEW hnsw_index_stats IS 
'Monitoring view for HNSW vector index sizes and stats.';

-- ============================================
-- 10. Best Practices Configuration
-- ============================================

-- Set default ef_search for good recall/speed balance
-- Can be overridden per-session with SET hnsw.ef_search = N
ALTER DATABASE current SET hnsw.ef_search = 64;

COMMENT ON FUNCTION search_team_memories_by_vector IS 
'HNSW search for team memories. Default ef_search=64. Increase for better recall, decrease for speed.';
