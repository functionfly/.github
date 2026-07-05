DROP TABLE IF EXISTS function_artifact_migration_cursor;

DROP INDEX IF EXISTS idx_rf_code_storage_backend_migrated;
ALTER TABLE registry_functions
    DROP COLUMN IF EXISTS code_content_hash,
    DROP COLUMN IF EXISTS code_storage_key,
    DROP COLUMN IF EXISTS code_storage_backend;

DROP INDEX IF EXISTS idx_rfv_storage_backend_migrated;
DROP INDEX IF EXISTS idx_rfv_artifact_hash;
ALTER TABLE registry_function_versions
    DROP COLUMN IF EXISTS artifact_hash,
    DROP COLUMN IF EXISTS readme_storage_key,
    DROP COLUMN IF EXISTS source_storage_key,
    DROP COLUMN IF EXISTS storage_key,
    DROP COLUMN IF EXISTS storage_backend;
