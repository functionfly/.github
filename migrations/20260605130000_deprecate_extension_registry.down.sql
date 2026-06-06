-- Rollback for 20260605130000_deprecate_extension_registry.
-- Safe to run only BEFORE the table is dropped in the follow-up
-- migration (30-day window). After that, the table itself is gone
-- and this rollback is a no-op for the table.

BEGIN;

DROP INDEX IF EXISTS idx_extension_registry_deprecated_at;

COMMENT ON TABLE extension_registry IS NULL;
COMMENT ON COLUMN extension_registry.deprecated_at IS NULL;

ALTER TABLE extension_registry
    DROP COLUMN IF EXISTS deprecated_at;

COMMIT;
