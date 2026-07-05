-- Drop the legacy artifact storage columns from Postgres. Apply only after
-- the migration worker (cmd/migrate-function-artifacts --watch) has reported
-- 0 rows where storage_backend='db' AND wasm_binary/source_code/readme IS NOT NULL.
--
-- A row-count guard is included so this migration is a no-op while any
-- legacy data remains — preventing accidental data loss.

DO $$
DECLARE
    leftover BIGINT;
BEGIN
    SELECT count(*) INTO leftover
    FROM registry_function_versions
    WHERE storage_backend = 'db'
      AND (wasm_binary IS NOT NULL OR source_code IS NOT NULL OR readme IS NOT NULL);
    IF leftover > 0 THEN
        RAISE EXCEPTION 'cannot drop legacy columns: % row(s) still in registry_function_versions (storage_backend=db). Run cmd/migrate-function-artifacts --watch first.', leftover;
    END IF;

    SELECT count(*) INTO leftover
    FROM registry_functions
    WHERE code_storage_backend = 'db'
      AND code IS NOT NULL
      AND length(code) > 0;
    IF leftover > 0 THEN
        RAISE EXCEPTION 'cannot drop legacy code column: % row(s) still in registry_functions (code_storage_backend=db). Run cmd/migrate-function-artifacts --watch first.', leftover;
    END IF;
END $$;

DROP INDEX IF EXISTS idx_rfv_storage_backend_migrated;
ALTER TABLE registry_function_versions
    DROP COLUMN IF EXISTS readme,
    DROP COLUMN IF EXISTS source_code,
    DROP COLUMN IF EXISTS wasm_binary;

ALTER TABLE registry_functions
    DROP COLUMN IF EXISTS code;
