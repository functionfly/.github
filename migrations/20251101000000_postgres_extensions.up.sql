-- PostgreSQL extensions required or recommended by the application.
-- Run this migration first so later migrations can use vector, pg_trgm, etc.
--
-- Required (standard contrib):
--   pg_trgm  - similarity() for recommendations (000050)
--   uuid-ossp - some migrations expect it; core has gen_random_uuid() too
--
-- Optional (created when available; migrations use BYTEA/stub when missing):
--   vector (pgvector) - agent_memories.embedding, function_embeddings.embedding
--   pgcrypto - encryption (internal/security/database.go)
--   pg_stat_statements - db_query_performance monitoring view
--   unaccent - accent-insensitive full-text search (registry, search)
--
-- Install pgvector: apt install postgresql-17-pgvector (PGDG) or brew install pgvector
-- See docs/LOCAL_POSTGRES_17.md for migrating local Debian DB to PG 17 + extensions.

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

-- pgcrypto: used by app for encryption; skip if not available
DO $$
BEGIN
  CREATE EXTENSION IF NOT EXISTS "pgcrypto";
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

-- unaccent: improves full-text search (accent-insensitive); skip if not available
DO $$
BEGIN
  CREATE EXTENSION IF NOT EXISTS unaccent;
EXCEPTION
  WHEN OTHERS THEN
    NULL;
END
$$;
