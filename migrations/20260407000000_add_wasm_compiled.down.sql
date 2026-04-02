-- Rollback: remove AOT-compiled WASM module storage

ALTER TABLE registry_function_versions
DROP COLUMN IF EXISTS wasm_compiled;
