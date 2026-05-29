-- Drop search optimization indexes
DROP INDEX IF EXISTS idx_registry_functions_name_trgm;
DROP INDEX IF EXISTS idx_registry_functions_title_trgm;
DROP INDEX IF EXISTS idx_registry_functions_description_trgm;
DROP INDEX IF EXISTS idx_registry_functions_author_name_public;
DROP INDEX IF EXISTS idx_registry_functions_trust_popularity_public;