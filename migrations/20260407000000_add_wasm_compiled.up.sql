-- Add AOT-compiled WASM module storage to registry function versions.
-- This column stores pre-compiled .cwasm bytes so the runtime can deserialize
-- in ~0.1ms instead of recompiling on every cold start.

ALTER TABLE registry_function_versions
ADD COLUMN IF NOT EXISTS wasm_compiled BYTEA;

COMMENT ON COLUMN registry_function_versions.wasm_compiled IS 'AOT-compiled Wasmtime module bytes (.cwasm) for near-instant cold starts';
