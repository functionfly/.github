-- Rollback determinism classification fields
DROP INDEX IF EXISTS idx_registry_function_versions_cacheable;

ALTER TABLE registry_function_versions
DROP COLUMN IF EXISTS idempotent,
DROP COLUMN IF EXISTS side_effects;
