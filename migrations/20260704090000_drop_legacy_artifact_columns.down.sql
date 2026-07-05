ALTER TABLE registry_function_versions
    ADD COLUMN IF NOT EXISTS wasm_binary BYTEA,
    ADD COLUMN IF NOT EXISTS source_code TEXT,
    ADD COLUMN IF NOT EXISTS readme       TEXT;

ALTER TABLE registry_functions
    ADD COLUMN IF NOT EXISTS code TEXT;

CREATE INDEX IF NOT EXISTS idx_rfv_storage_backend_migrated
    ON registry_function_versions(storage_backend)
    WHERE storage_backend = 'db'
      AND (wasm_binary IS NOT NULL OR source_code IS NOT NULL OR readme IS NOT NULL);
