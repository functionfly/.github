-- Rollback: agent_mcp_servers

BEGIN;

DROP TABLE IF EXISTS agent_mcp_servers;

COMMIT;
