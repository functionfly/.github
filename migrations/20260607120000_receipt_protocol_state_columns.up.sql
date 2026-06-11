-- P1: Add protocol, state, parent_task_id, fallback_chain to registry_executions_public.
-- A2A Tasks live in the SAME table as MCP receipts. The protocol column is the
-- discriminator; the state column holds the A2A state machine.

ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS protocol TEXT NOT NULL DEFAULT 'mcp'
    CHECK (protocol IN ('mcp', 'a2a'));

ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS state TEXT NOT NULL DEFAULT 'completed'
    CHECK (state IN (
      'submitted', 'working', 'input-required',
      'completed', 'failed', 'canceled'
    ));

ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS parent_task_id UUID NULL
    REFERENCES registry_executions_public(id) ON DELETE SET NULL;

ALTER TABLE registry_executions_public
  ADD COLUMN IF NOT EXISTS fallback_chain TEXT[] NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_exec_public_protocol
  ON registry_executions_public(protocol, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exec_public_state
  ON registry_executions_public(protocol, state)
  WHERE state IN ('submitted', 'working', 'input-required');

CREATE INDEX IF NOT EXISTS idx_exec_public_parent
  ON registry_executions_public(parent_task_id)
  WHERE parent_task_id IS NOT NULL;
