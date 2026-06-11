-- Add composite index on registry_functions for author/name lookups
-- This optimizes the common query pattern: GetFunctionByAuthorName
CREATE INDEX IF NOT EXISTS idx_registry_functions_author_name 
ON registry_functions (author, name);
