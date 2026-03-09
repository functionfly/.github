-- Drop extensions in reverse order (dependencies first).
-- Only drop if we own them; skip if they were pre-existing.

DROP EXTENSION IF EXISTS pg_stat_statements;
DROP EXTENSION IF EXISTS vector;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS "pg_trgm";
