-- Add readme column to registry_function_versions for function documentation
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS readme TEXT;

-- Index for readme presence queries
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_has_readme ON registry_function_versions (((readme IS NOT NULL) AND (readme <> '')));