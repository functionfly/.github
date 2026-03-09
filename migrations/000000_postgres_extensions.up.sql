-- PostgreSQL extensions required or recommended by the application.
-- Run this migration first so later migrations can use vector, pg_trgm, etc.
--
-- Required (standard contrib):
--   pg_trgm  - similarity() for recommendations (000050)
--   uuid-ossp - some migrations expect it; core has gen_random_uuid() too
--
-- Optional (created when available; migrations use BYTEA/stub when missing):
--   vector (pgvector) - agent_memories.embedding, function_embeddings.embedding
--   pg_stat_statements - db_query_performance monitoring view
--
-- Install pgvector: apt install postgresql-16-pgvector  or  brew install pgvector

CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Optional: skip if extension package not installed (e.g. no pgvector on server)
DO $$
BEGIN
  CREATE EXTENSION IF NOT EXISTS vector;
EXCEPTION
  WHEN OTHERS THEN
    NULL;
END
$$;

DO $$
BEGIN
  CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
EXCEPTION
  WHEN OTHERS THEN
    NULL;
END
$$;
