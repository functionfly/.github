-- Registry Executions Public Table
-- Stores shareable executions for the playground/replay feature

CREATE TABLE IF NOT EXISTS registry_executions_public (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    public_id VARCHAR(24) UNIQUE NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id),
    version VARCHAR(20) NOT NULL,
    input_json JSONB NOT NULL,
    output_json JSONB NOT NULL,
    duration_ms INTEGER NOT NULL,
    cached BOOLEAN DEFAULT false,
    shareable BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_registry_executions_public_function_id ON registry_executions_public(function_id);
CREATE INDEX IF NOT EXISTS idx_registry_executions_public_created_at ON registry_executions_public(created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_registry_executions_public_public_id ON registry_executions_public(public_id);