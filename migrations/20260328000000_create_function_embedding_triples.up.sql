-- FlyEmbed triple-vector ColBERT embeddings
-- Each function gets three specialized 512-dim vectors: contract, semantic, code
CREATE TABLE IF NOT EXISTS function_embedding_triples (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Three specialized vectors (512 dims each)
    contract_embedding vector(512),
    semantic_embedding vector(512),
    code_embedding vector(512),

    -- Source texts for debugging/re-embedding
    contract_text TEXT,
    semantic_text TEXT,
    code_text TEXT,

    -- Metadata
    embedding_model VARCHAR(100) NOT NULL DEFAULT 'flyembed-v1',
    embedding_version INTEGER NOT NULL DEFAULT 1,
    computed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_triple_embedding UNIQUE (function_id)
);

-- HNSW indexes for each vector column (cosine distance)
CREATE INDEX idx_triple_contract_hnsw ON function_embedding_triples
    USING hnsw (contract_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
CREATE INDEX idx_triple_semantic_hnsw ON function_embedding_triples
    USING hnsw (semantic_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
CREATE INDEX idx_triple_code_hnsw ON function_embedding_triples
    USING hnsw (code_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
