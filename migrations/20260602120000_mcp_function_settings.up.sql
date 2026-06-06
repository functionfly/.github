-- 20260602120000_mcp_function_settings.up.sql
--
-- MCP (Model Context Protocol) function settings: per-function configuration
-- for exposing functions through the MCP Function Registry.
--
-- A row in this table makes a function "MCP-enabled" and visible to MCP clients
-- (Claude Desktop, Cursor, Continue, etc.) via the JSON-RPC 2.0 transport at
--   GET  /v1/mcp/manifest
--   POST /v1/mcp            (initialize / tools/list / tools/call)
--   GET  /v1/mcp/tools      (public, cacheable, SEO-indexable)
--
-- This migration is idempotent (CREATE TABLE / INDEX IF NOT EXISTS) so it is
-- safe to run on a database that may have partial state.

BEGIN;

CREATE TABLE IF NOT EXISTS registry_function_mcp_settings (
    function_id          UUID         PRIMARY KEY REFERENCES registry_functions(id) ON DELETE CASCADE,
    enabled              BOOLEAN      NOT NULL DEFAULT false,
    transports           TEXT[]       NOT NULL DEFAULT ARRAY['streamable-http']::TEXT[],
    expose_input_schema  BOOLEAN      NOT NULL DEFAULT true,
    expose_output_schema BOOLEAN      NOT NULL DEFAULT false,
    tool_name_override   TEXT         CHECK (tool_name_override IS NULL OR (
                                length(tool_name_override) BETWEEN 1 AND 64
                                AND tool_name_override ~ '^[a-zA-Z0-9_-]+$'
                           )),
    rate_limit_per_min   INTEGER      NOT NULL DEFAULT 60 CHECK (rate_limit_per_min BETWEEN 1 AND 10000),
    allowlist_origins    TEXT[]       NOT NULL DEFAULT ARRAY[]::TEXT[],
    verified_mcp         BOOLEAN      NOT NULL DEFAULT false,
    verified_at          TIMESTAMPTZ,
    verified_by          UUID         REFERENCES users(id) ON DELETE SET NULL,
    enabled_at           TIMESTAMPTZ,
    enabled_by           UUID         REFERENCES users(id) ON DELETE SET NULL,
    last_invoked_at      TIMESTAMPTZ,
    invocation_count     BIGINT       NOT NULL DEFAULT 0 CHECK (invocation_count >= 0),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),

    -- Enforce a sane transport set
    CONSTRAINT registry_function_mcp_settings_transports_nonempty
        CHECK (array_length(transports, 1) >= 1)
);

-- Index for the hot path: "give me all MCP-enabled public functions, paginated by id"
CREATE INDEX IF NOT EXISTS idx_mcp_settings_enabled_function
    ON registry_function_mcp_settings(function_id)
    WHERE enabled = true;

-- Index for admin dashboards and search
CREATE INDEX IF NOT EXISTS idx_mcp_settings_transports_gin
    ON registry_function_mcp_settings USING GIN(transports);

CREATE INDEX IF NOT EXISTS idx_mcp_settings_verified
    ON registry_function_mcp_settings(verified_mcp)
    WHERE verified_mcp = true;

-- updated_at trigger (inlined so this migration is self-contained)
CREATE OR REPLACE FUNCTION trg_mcp_settings_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_mcp_settings_updated_at ON registry_function_mcp_settings;
CREATE TRIGGER trg_mcp_settings_updated_at
    BEFORE UPDATE ON registry_function_mcp_settings
    FOR EACH ROW EXECUTE FUNCTION trg_mcp_settings_set_updated_at();

COMMIT;
