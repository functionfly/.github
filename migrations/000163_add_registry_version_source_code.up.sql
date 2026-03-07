-- Add source_code and updated_at for lazy bundling and GORM compatibility
-- +migrate Up
ALTER TABLE registry_function_versions
ADD COLUMN IF NOT EXISTS source_code TEXT,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

COMMENT ON COLUMN registry_function_versions.source_code IS 'Source code for lazy WASM bundling at first execution';
