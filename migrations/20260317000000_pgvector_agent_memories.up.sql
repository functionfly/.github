-- Enable pgvector for agent_memories: use vector(1536) and HNSW index for cosine similarity.
-- Requires: Postgres with pgvector (apt install postgresql-16-pgvector or image pgvector/pgvector:pg16).
-- If the vector extension is not available, this migration no-ops (agent_memories keeps BYTEA).

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

-- Convert agent_memories.embedding from BYTEA to vector(1536) when extension exists.
-- Existing BYTEA values are not converted (USING NULL); app can re-embed if needed.
DO $$
DECLARE
  col_type text;
BEGIN
  SELECT data_type INTO col_type
  FROM information_schema.columns
  WHERE table_schema = 'public' AND table_name = 'agent_memories' AND column_name = 'embedding';

  IF col_type = 'bytea' AND EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector') THEN
    ALTER TABLE agent_memories
      ALTER COLUMN embedding TYPE vector(1536) USING NULL;
  END IF;
EXCEPTION
  WHEN OTHERS THEN
    NULL;
END
$$;

-- HNSW index for cosine distance (<=>) only when embedding is vector type.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_attribute a
    JOIN pg_type t ON a.atttypid = t.oid
    WHERE a.attrelid = 'agent_memories'::regclass AND a.attname = 'embedding' AND t.typname = 'vector'
  ) THEN
    CREATE INDEX IF NOT EXISTS idx_agent_memories_embedding_hnsw
    ON agent_memories USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
  END IF;
EXCEPTION
  WHEN OTHERS THEN NULL;
END
$$;
