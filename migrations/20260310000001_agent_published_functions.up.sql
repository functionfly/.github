-- Agent published functions: tracks functions published by the factory agent to the registry
CREATE TABLE IF NOT EXISTS agent_published_functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL,
    generated_code_id UUID NOT NULL,
    function_id VARCHAR(255),
    registry_function_id UUID,
    author VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    title VARCHAR(255),
    description TEXT,
    category VARCHAR(255),
    tags TEXT[],
    is_public BOOLEAN NOT NULL DEFAULT false,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    error_message TEXT,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_published_functions_agent_id ON agent_published_functions(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_published_functions_status ON agent_published_functions(status);
CREATE INDEX IF NOT EXISTS idx_agent_published_functions_created_at ON agent_published_functions(created_at);
