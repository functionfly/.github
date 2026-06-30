ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS sandbox_tier VARCHAR(20) DEFAULT 'wasm';
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS sandbox_config JSONB DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_registry_function_versions_sandbox_tier ON registry_function_versions(sandbox_tier);

ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS default_sandbox_tier VARCHAR(20) DEFAULT 'wasm';
