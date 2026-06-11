-- Remove WASM storage from registry function versions
-- +migrate Down

-- Drop index first
DROP INDEX IF EXISTS idx_registry_function_versions_wasm;

-- Remove the new columns
ALTER TABLE registry_function_versions
DROP COLUMN IF EXISTS wasm_binary,
DROP COLUMN IF EXISTS source_hash,
DROP COLUMN IF EXISTS bundle_size;