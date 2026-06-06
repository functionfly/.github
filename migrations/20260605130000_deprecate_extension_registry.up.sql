-- Migration: deprecate_extension_registry
-- Marks the legacy extension_registry table and its rows as deprecated.
-- The table, its repository, and its HTTP handler continue to function
-- but every response carries Sunset / Deprecation / X-Deprecated-Use
-- headers (see internal/api/handlers/studio/extensions_handler.go).
-- Removal is scheduled for 2026-07-06 in a follow-up migration.
--
-- See docs/PLUGIN_MIGRATION.md for the full client migration guide.

BEGIN;

ALTER TABLE extension_registry
    ADD COLUMN IF NOT EXISTS deprecated_at TIMESTAMPTZ;

UPDATE extension_registry
SET deprecated_at = COALESCE(deprecated_at, NOW())
WHERE deprecated_at IS NULL;

-- Add a partial index so the deprecation sweep can quickly find any row
-- that somehow gets re-inserted without deprecated_at set. Defensive:
-- nothing in the codebase writes to extension_registry anymore, but
-- the table is still readable.
CREATE INDEX IF NOT EXISTS idx_extension_registry_deprecated_at
    ON extension_registry(deprecated_at)
    WHERE deprecated_at IS NULL;

-- A comment on the table is the single most visible "this is gone soon"
-- signal for anyone reading the schema in psql or a migration tool.
COMMENT ON TABLE extension_registry IS
    'DEPRECATED 2026-06-06. Sunset 2026-07-06. Use the plugins table instead. See docs/PLUGIN_MIGRATION.md.';

-- A comment on the deprecated_at column explains the new contract.
COMMENT ON COLUMN extension_registry.deprecated_at IS
    'Timestamp the row was marked deprecated. NULL means the row predates the deprecation migration and was back-filled by 20260605130000.';

COMMIT;
