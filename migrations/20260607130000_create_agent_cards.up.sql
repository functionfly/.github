-- P6: Agent Card Directory table.
-- Parallel to mcp_settings; stores A2A agent cards for discovery and routing.

CREATE TABLE IF NOT EXISTS agent_cards (
  id              TEXT PRIMARY KEY,
  version         TEXT NOT NULL DEFAULT '1.0',
  name            TEXT NOT NULL,
  description     TEXT,
  url             TEXT,
  protocol_version TEXT NOT NULL DEFAULT '0.3.0',
  capabilities    TEXT[] NOT NULL DEFAULT '{}',
  skills          JSONB NOT NULL DEFAULT '[]',
  auth_schemes    TEXT[] NOT NULL DEFAULT '{}',
  input_modes     TEXT[] NOT NULL DEFAULT '{application/json}',
  output_modes    TEXT[] NOT NULL DEFAULT '{application/json}',
  trust_score     DOUBLE PRECISION NOT NULL DEFAULT 0,
  peer_jwks_url   TEXT,
  published_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for browsing agent cards by trust score.
CREATE INDEX IF NOT EXISTS idx_agent_cards_trust
  ON agent_cards(trust_score DESC, published_at DESC);

-- Index for filtering by capability.
CREATE INDEX IF NOT EXISTS idx_agent_cards_capabilities
  ON agent_cards USING GIN(capabilities);

-- Index for filtering by skill ID (JSONB path).
CREATE INDEX IF NOT EXISTS idx_agent_cards_skills
  ON agent_cards USING GIN(skills);
