-- Migration: Add embed_config to registry_functions and embed_origin to registry_function_executions
-- Feature: Function Embeds (Feature 3)

-- Add embed_config JSONB column to registry_functions
ALTER TABLE registry_functions
    ADD COLUMN IF NOT EXISTS embed_config JSONB DEFAULT NULL;

COMMENT ON COLUMN registry_functions.embed_config IS
    'Per-function embed configuration: enabled, allowed_origins, require_api_key, ui_enabled, ui_theme, rate_limit_per_hour';

-- Add embed_origin TEXT column to registry_function_executions
ALTER TABLE registry_function_executions
    ADD COLUMN IF NOT EXISTS embed_origin TEXT DEFAULT NULL;

COMMENT ON COLUMN registry_function_executions.embed_origin IS
    'Origin domain that triggered the embed execution (for analytics)';

-- Index for analytics queries on embed_origin
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_embed_origin
    ON registry_function_executions (embed_origin)
    WHERE embed_origin IS NOT NULL;
