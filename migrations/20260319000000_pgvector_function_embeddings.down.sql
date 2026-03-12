-- Revert function_embeddings pgvector: drop HNSW index and column type back to BYTEA.

DROP INDEX IF EXISTS idx_function_embeddings_embedding_hnsw;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = 'function_embeddings' AND column_name = 'embedding')
     AND (SELECT data_type FROM information_schema.columns
          WHERE table_schema = 'public' AND table_name = 'function_embeddings' AND column_name = 'embedding') = 'USER-DEFINED' THEN
    ALTER TABLE function_embeddings ALTER COLUMN embedding TYPE bytea USING NULL;
  END IF;
EXCEPTION
  WHEN OTHERS THEN
    NULL;
END
$$;
