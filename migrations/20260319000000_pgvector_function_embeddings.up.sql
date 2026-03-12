-- Enable pgvector for function_embeddings: use vector(1536) and HNSW index for cosine similarity.
-- Same pattern as agent_memories (see docs/PGVECTOR_SETUP.md).
-- If the vector extension is not available, this migration no-ops.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
    CREATE EXTENSION IF NOT EXISTS vector;
  END IF;
EXCEPTION
  WHEN OTHERS THEN
    NULL;
END
$$;

-- Convert function_embeddings.embedding from BYTEA to vector(1536) when extension exists.
DO $$
DECLARE
  col_type text;
BEGIN
  SELECT data_type INTO col_type
  FROM information_schema.columns
  WHERE table_schema = 'public' AND table_name = 'function_embeddings' AND column_name = 'embedding';

  IF col_type = 'bytea' AND EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
    ALTER TABLE function_embeddings
      ALTER COLUMN embedding TYPE vector(1536) USING NULL;
  END IF;
EXCEPTION
  WHEN OTHERS THEN
    NULL;
END
$$;

-- HNSW index for cosine distance (<=>) for similarity search.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_attribute a
    JOIN pg_type t ON a.atttypid = t.oid
    WHERE a.attrelid = 'function_embeddings'::regclass AND a.attname = 'embedding' AND t.typname = 'vector'
  ) THEN
    CREATE INDEX IF NOT EXISTS idx_function_embeddings_embedding_hnsw
    ON function_embeddings USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
  END IF;
EXCEPTION
  WHEN OTHERS THEN NULL;
END
$$;
