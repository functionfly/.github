-- Add settings JSONB to registry_functions for per-function settings (e.g. custom_domains)
ALTER TABLE registry_functions
ADD COLUMN IF NOT EXISTS settings JSONB DEFAULT '{}';

COMMENT ON COLUMN registry_functions.settings IS 'Per-function settings: custom_domains (array), etc.';
