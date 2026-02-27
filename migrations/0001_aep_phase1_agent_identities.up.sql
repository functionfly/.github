-- AEP Phase 1: Agent Identity and Quota Configuration
-- Migration: 0001_aep_phase1_agent_identities

-- Agent identities table
CREATE TABLE IF NOT EXISTS agent_identities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    agent_id    TEXT NOT NULL UNIQUE,  -- "org/agent-name"
    name        TEXT NOT NULL,
    description TEXT,
    plan_tier   TEXT NOT NULL DEFAULT 'agent_starter',
    status      TEXT NOT NULL DEFAULT 'active',  -- active | suspended | deleted
    api_key_hash TEXT,  -- hashed API key for agent auth
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_identities_tenant_id ON agent_identities(tenant_id);
CREATE INDEX IF NOT EXISTS idx_agent_identities_status ON agent_identities(status);

-- Agent quota configurations table
CREATE TABLE IF NOT EXISTS agent_quota_configs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id                TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    max_calls_per_minute    INT NOT NULL DEFAULT 100,
    max_calls_per_day       INT NOT NULL DEFAULT 16667,
    max_state_writes_per_hr INT NOT NULL DEFAULT 1000,
    max_cost_per_execution  DECIMAL(10,6) NOT NULL DEFAULT 0.01,
    max_daily_spend_usd     DECIMAL(10,2) NOT NULL DEFAULT 5.00,
    allowed_functions       TEXT[],  -- NULL = all allowed
    forbidden_functions     TEXT[],
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_quota_configs_agent_id ON agent_quota_configs(agent_id);
