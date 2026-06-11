-- P6: Agent Function Registry and Execution Tracking
-- Extends registry_functions with agent-specific metadata and execution logs

-- Agent function registry (extends registry_functions)
CREATE TABLE IF NOT EXISTS agent_functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL, -- 'search', 'browser', 'file', 'data', 'compute', 'communication', 'workflow', 'memory', 'assure', 'validate', 'simulate', 'observe', 'learn', 'agent_mgmt', 'capability'
    capabilities TEXT[] NOT NULL DEFAULT '{}', -- ['web_access', 'ssl', 'javascript', 'filesystem', 'smtp', 'compliance', 'simulation']
    input_schema JSONB NOT NULL DEFAULT '{}',
    output_schema JSONB NOT NULL DEFAULT '{}',
    is_verified BOOLEAN DEFAULT false,
    is_exclusive BOOLEAN DEFAULT false, -- FunctionFly exclusive function
    max_concurrency INTEGER DEFAULT 10,
    rate_limit_rpm INTEGER DEFAULT 60,
    pricing_model JSONB NOT NULL DEFAULT '{"type": "per_call", "price": 0.001}', -- {'type': 'per_call', 'price': 0.001}
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(function_id)
);

-- Function execution logs for agents
CREATE TABLE IF NOT EXISTS agent_function_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    session_id UUID,
    input JSONB NOT NULL DEFAULT '{}',
    output JSONB,
    error TEXT,
    duration_ms INTEGER,
    cost_usd DECIMAL(10, 6) DEFAULT 0,
    trace_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Index for analytics by agent
CREATE INDEX IF NOT EXISTS idx_agent_function_executions_agent ON agent_function_executions(agent_id, created_at DESC);

-- Index for analytics by function
CREATE INDEX IF NOT EXISTS idx_agent_function_executions_function ON agent_function_executions(function_id, created_at DESC);

-- Index for session-based queries
CREATE INDEX IF NOT EXISTS idx_agent_function_executions_session ON agent_function_executions(session_id, created_at DESC);

-- Agent function policies (per-agent per-function allowlist)
CREATE TABLE IF NOT EXISTS agent_function_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    allowed BOOLEAN DEFAULT true,
    max_calls_per_day INTEGER,
    max_cost_per_call_usd DECIMAL(10, 6),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(agent_id, function_id)
);

-- Index for policy lookups
CREATE INDEX IF NOT EXISTS idx_agent_function_policies_agent ON agent_function_policies(agent_id);

-- FunctionFly exclusive function categories view for discovery
CREATE OR REPLACE VIEW agent_exclusive_functions AS
SELECT
    af.id,
    af.function_id,
    rf.author,
    rf.name,
    rf.title as display_name,
    rf.description,
    af.category,
    af.capabilities,
    af.input_schema,
    af.output_schema,
    af.pricing_model,
    af.is_verified,
    af.max_concurrency,
    af.rate_limit_rpm,
    af.created_at
FROM agent_functions af
JOIN registry_functions rf ON rf.id = af.function_id
WHERE af.is_exclusive = true;

-- Function categories enum for validation
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'agent_function_category') THEN
        CREATE TYPE agent_function_category AS ENUM (
            'search', 'browser', 'file', 'data', 'compute',
            'communication', 'workflow', 'memory',
            'assure', 'validate', 'simulate', 'observe', 'learn',
            'agent_mgmt', 'capability'
        );
    END IF;
END$$;