-- AEP Phase 2: Behavioral Policies, Execution Records, Sessions
-- Migration: 0002_aep_phase2_execution_tracking

-- Agent behavioral policies table
CREATE TABLE IF NOT EXISTS agent_behavioral_policies (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id                TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    max_execution_depth     INT NOT NULL DEFAULT 10,
    max_recursion_depth     INT NOT NULL DEFAULT 3,
    max_wall_time_ms        INT NOT NULL DEFAULT 300000,  -- 5 minutes
    max_memory_growth_mb    INT NOT NULL DEFAULT 512,
    forbidden_functions     TEXT[],
    deterministic_only      BOOLEAN NOT NULL DEFAULT FALSE,
    allowed_capabilities    TEXT[],
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_behavioral_policies_agent_id ON agent_behavioral_policies(agent_id);

-- Agent execution records table (hot metadata only)
CREATE TABLE IF NOT EXISTS agent_execution_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id        TEXT NOT NULL,
    tenant_id       UUID NOT NULL,
    function_id     UUID NOT NULL,
    execution_id    TEXT NOT NULL,
    session_id      TEXT,
    call_depth      INT NOT NULL DEFAULT 0,
    cost_usd        DECIMAL(10,6) NOT NULL DEFAULT 0,
    latency_ms      INT NOT NULL,
    outcome         TEXT NOT NULL,  -- success | error | timeout | policy_violation
    error_code      TEXT,
    object_key      TEXT,  -- pointer to full record in object storage
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_exec_agent_id ON agent_execution_records(agent_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_agent_exec_tenant_id ON agent_execution_records(tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_agent_exec_session_id ON agent_execution_records(session_id) WHERE session_id IS NOT NULL;

-- Agent sessions table
CREATE TABLE IF NOT EXISTS agent_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      TEXT NOT NULL UNIQUE,
    agent_id        TEXT NOT NULL,
    tenant_id       UUID NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active',  -- active | completed | terminated
    call_count      INT NOT NULL DEFAULT 0,
    total_cost_usd  DECIMAL(10,6) NOT NULL DEFAULT 0,
    call_graph      JSONB,  -- stored as compressed call graph
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at        TIMESTAMPTZ,
    object_key      TEXT  -- pointer to full session data in object storage
);

CREATE INDEX IF NOT EXISTS idx_agent_sessions_agent_id ON agent_sessions(agent_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_sessions_tenant_id ON agent_sessions(tenant_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_sessions_status ON agent_sessions(status) WHERE status = 'active';
