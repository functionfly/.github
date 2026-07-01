-- Migration: agent_mcp_servers
-- Purpose: Store MCP servers connected to agents for tool execution

BEGIN;

CREATE TABLE IF NOT EXISTS agent_mcp_servers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id        VARCHAR(255) NOT NULL,
    tenant_id       UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    url             TEXT NOT NULL,
    transport       VARCHAR(20) NOT NULL DEFAULT 'streamable-http',
    description     TEXT NOT NULL DEFAULT '',
    enabled         BOOLEAN NOT NULL DEFAULT true,
    headers         JSONB NOT NULL DEFAULT '{}',
    tool_count      INT NOT NULL DEFAULT 0,
    last_connected_at TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agent_mcp_servers_agent_id ON agent_mcp_servers(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_mcp_servers_tenant_id ON agent_mcp_servers(tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_mcp_servers_agent_url ON agent_mcp_servers(agent_id, url);

COMMIT;
