-- Move user-uploaded function artifacts from Postgres bytea/text to object storage.
--
-- This migration adds the metadata columns required for content-addressed
-- lookup. Legacy columns (wasm_binary, source_code, readme, code) are kept
-- during the dual-read cutover window and nullified by the background
-- migration worker (cmd/migrate-function-artifacts). A follow-up migration
-- drops them entirely once the cutover window closes.

ALTER TABLE registry_function_versions
    ADD COLUMN IF NOT EXISTS storage_backend     VARCHAR(20) NOT NULL DEFAULT 'db',
    ADD COLUMN IF NOT EXISTS storage_key         TEXT,
    ADD COLUMN IF NOT EXISTS source_storage_key  TEXT,
    ADD COLUMN IF NOT EXISTS readme_storage_key  TEXT,
    ADD COLUMN IF NOT EXISTS artifact_hash       VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_rfv_artifact_hash
    ON registry_function_versions(artifact_hash)
    WHERE artifact_hash IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_rfv_storage_backend_migrated
    ON registry_function_versions(storage_backend)
    WHERE storage_backend = 'db'
      AND (wasm_binary IS NOT NULL OR source_code IS NOT NULL OR readme IS NOT NULL);

ALTER TABLE registry_functions
    ADD COLUMN IF NOT EXISTS code_storage_backend VARCHAR(20) NOT NULL DEFAULT 'db',
    ADD COLUMN IF NOT EXISTS code_storage_key     TEXT,
    ADD COLUMN IF NOT EXISTS code_content_hash    VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_rf_code_storage_backend_migrated
    ON registry_functions(code_storage_backend)
    WHERE code_storage_backend = 'db'
      AND code IS NOT NULL
      AND length(code) > 0;

-- Resumable cursor tracking for the migration worker.
CREATE TABLE IF NOT EXISTS function_artifact_migration_cursor (
    id            SMALLINT PRIMARY KEY DEFAULT 1,
    last_version_id UUID,
    last_function_id UUID,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (id = 1)
);

INSERT INTO function_artifact_migration_cursor (id) VALUES (1) ON CONFLICT DO NOTHING;
