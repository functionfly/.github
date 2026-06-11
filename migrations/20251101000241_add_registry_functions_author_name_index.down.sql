-- Drop composite index on registry_functions for author/name lookups
DROP INDEX IF EXISTS idx_registry_functions_author_name;
