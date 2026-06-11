-- Migration: Agent Observability (Atlas integration)
-- Records AI agent decision-making with Atlas Memory Engine for replay and debugging

-- Agent observability runs: links Atlas runs to FunctionFly entities
-- Stores ONLY metadata; events live in Atlas
CREATE TABLE IF NOT EXISTS agent_observability_runs (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID NOT NULL REFERENCES tenants(id),
    atlas_tenant_id      TEXT NOT NULL,
    atlas_run_id         TEXT NOT NULL,
    agent_id             TEXT NOT NULL,
    agent_type           TEXT NOT NULL CHECK (agent_type IN ('flymind', 'agent', 'workflow', 'team')),
    span_id              TEXT,
    parent_span_id       TEXT,
    status               TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'completed', 'failed')),
    total_cost_usd      DECIMAL(12, 6) DEFAULT 0,
    total_input_tokens   INT DEFAULT 0,
    total_output_tokens  INT DEFAULT 0,
    event_count          INT DEFAULT 0,
    error_count          INT DEFAULT 0,
    tool_call_count      INT DEFAULT 0,
    started_at           TIMESTAMPTZ NOT NULL,
    ended_at             TIMESTAMPTZ,
    last_event_at        TIMESTAMPTZ,
    metadata             JSONB DEFAULT '{}',
    created_at           TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_observability_runs_tenant ON agent_observability_runs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_observability_runs_agent ON agent_observability_runs(tenant_id, agent_id);
CREATE INDEX IF NOT EXISTS idx_observability_runs_atlas_run ON agent_observability_runs(atlas_run_id);
CREATE INDEX IF NOT EXISTS idx_observability_runs_status ON agent_observability_runs(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_observability_runs_started ON agent_observability_runs(started_at DESC);

-- Sampling configuration per tenant
CREATE TABLE IF NOT EXISTS agent_observability_config (
    tenant_id            UUID PRIMARY KEY REFERENCES tenants(id),
    atlas_tenant_id      TEXT NOT NULL,
    sampling_rate        DECIMAL(3, 2) DEFAULT 1.0 CHECK (sampling_rate >= 0 AND sampling_rate <= 1),
    trace_errors_only    BOOLEAN DEFAULT false,
    sample_head_percent  DECIMAL(5, 2) DEFAULT 100 CHECK (sample_head_percent >= 0 AND sample_head_percent <= 100),
    sample_tail_count    INT DEFAULT 10,
    retention_days       INT DEFAULT 90,
    is_active            BOOLEAN DEFAULT true,
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);