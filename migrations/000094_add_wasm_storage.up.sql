-- Add WASM storage to registry function versions
-- +migrate Up

-- Add WASM binary storage to function versions
ALTER TABLE registry_function_versions
ADD COLUMN wasm_binary BYTEA,
ADD COLUMN source_hash VARCHAR(64),
ADD COLUMN bundle_size INTEGER;

-- Add index for efficient WASM queries
CREATE INDEX idx_registry_function_versions_wasm ON registry_function_versions(wasm_binary) WHERE wasm_binary IS NOT NULL;

-- Add comment explaining the new columns
COMMENT ON COLUMN registry_function_versions.wasm_binary IS 'Compiled WebAssembly binary for sandbox execution';
COMMENT ON COLUMN registry_function_versions.source_hash IS 'SHA256 hash of source code used to generate WASM';
COMMENT ON COLUMN registry_function_versions.bundle_size IS 'Size of WASM binary in bytes';