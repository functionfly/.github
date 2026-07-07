-- Migration: Add registry_mcp_global_settings table for platform-wide MCP defaults
-- This table stores platform-wide MCP configuration defaults that apply across all functions

CREATE TABLE IF NOT EXISTS registry_mcp_global_settings (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),  -- Single-row config table
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ensure only one row exists
CREATE UNIQUE INDEX IF NOT EXISTS idx_mcp_global_settings_single_row ON registry_mcp_global_settings (id) WHERE id = 1;

-- Insert default configuration
INSERT INTO registry_mcp_global_settings (id, config)
VALUES (1, '{
    "default_transport": "streamable-http",
    "default_rate_limit": 60,
    "default_expose_input": true,
    "default_expose_output": false,
    "auto_add_to_registry": false,
    "require_verification": true,
    "public_listing": true,
    "cors_allowlist": [],
    "rate_limit_multiplier": 1
}')
ON CONFLICT (id) DO NOTHING;
