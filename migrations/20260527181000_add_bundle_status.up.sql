-- Add bundle_status column to track WASM bundling state
-- Default 'lazy' means no WASM binary at publish time; will be bundled at execution time.
-- 'pending' = eager bundling in progress (set asynchronously after publish).
-- 'bundled' = WASM binary successfully generated (either pre-compiled or eager-bundled).
-- 'failed' = eager bundling failed; execution will attempt lazy bundling.
ALTER TABLE registry_function_versions
ADD COLUMN IF NOT EXISTS bundle_status VARCHAR(20) DEFAULT 'lazy';

COMMENT ON COLUMN registry_function_versions.bundle_status IS 'WASM bundling state: pending|bundled|lazy|failed';
