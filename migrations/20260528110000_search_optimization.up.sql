-- Add trigram GIN indexes for fast ILIKE '%query%' searches
-- This dramatically speeds up wildcard text searches used in function discovery

-- Enable pg_trgm extension (required for trigram indexes)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Trigram index for name column (most selective for function search)
CREATE INDEX IF NOT EXISTS idx_registry_functions_name_trgm
ON registry_functions
USING gin (name gin_trgm_ops)
WHERE visibility = 'public';

-- Trigram index for title column
CREATE INDEX IF NOT EXISTS idx_registry_functions_title_trgm
ON registry_functions
USING gin (title gin_trgm_ops)
WHERE visibility = 'public';

-- Trigram index for description column (larger text, helps partial matches)
CREATE INDEX IF NOT EXISTS idx_registry_functions_description_trgm
ON registry_functions
USING gin (description gin_trgm_ops)
WHERE visibility = 'public';

-- Combined text search using search_vector (already has GIN tsvector index)
-- but add a specific index for prefix/wildcard matching on individual columns

-- Combined index for author+name lookups (common in function discovery)
CREATE INDEX IF NOT EXISTS idx_registry_functions_author_name_public
ON registry_functions (author, name)
WHERE visibility = 'public';

-- Add composite index for the search + sort pattern used in function discovery
CREATE INDEX IF NOT EXISTS idx_registry_functions_trust_popularity_public
ON registry_functions (trust_score DESC, popularity_score DESC, reliability_score DESC)
WHERE visibility = 'public';

-- Analyze to update statistics
ANALYZE registry_functions;