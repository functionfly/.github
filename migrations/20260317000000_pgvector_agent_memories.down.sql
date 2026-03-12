-- Remove pgvector index and revert agent_memories.embedding to BYTEA (optional; only if you need rollback).

DROP INDEX IF EXISTS idx_agent_memories_embedding_hnsw;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'agent_memories' AND column_name = 'embedding')
     AND (SELECT data_type FROM information_schema.columns
          WHERE table_schema = 'public' AND table_name = 'agent_memories' AND column_name = 'embedding') = 'USER-DEFINED' THEN
    ALTER TABLE agent_memories ALTER COLUMN embedding TYPE bytea USING NULL;
  END IF;
EXCEPTION
  WHEN OTHERS THEN
    NULL;
END
$$;
