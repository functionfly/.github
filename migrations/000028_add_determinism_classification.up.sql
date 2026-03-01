-- Add determinism classification fields to registry_function_versions
-- These fields enable intelligent caching and replay decisions

-- Add side_effects column: none | network | external_state
ALTER TABLE registry_function_versions
ADD COLUMN side_effects VARCHAR(20) DEFAULT 'none'
CHECK (side_effects IN ('none', 'network', 'external_state'));

-- Add idempotent column: true | false
ALTER TABLE registry_function_versions
ADD COLUMN idempotent BOOLEAN DEFAULT false;

-- Add index for efficient filtering of cacheable functions
-- Functions that are deterministic + no side effects + idempotent = safe to CDN cache
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_cacheable
ON registry_function_versions(deterministic, side_effects, idempotent)
WHERE deterministic = true AND side_effects = 'none' AND idempotent = true;

-- Add comments explaining the new columns
COMMENT ON COLUMN registry_function_versions.side_effects IS 'Side effects: none (no side effects), network (makes external calls), external_state (modifies external state)';
COMMENT ON COLUMN registry_function_versions.idempotent IS 'Whether the function is idempotent (safe to retry with same input)';

-- Update deterministic column comment to reflect its role in caching
COMMENT ON COLUMN registry_function_versions.deterministic IS 'Whether function is deterministic: same input always produces same output';
