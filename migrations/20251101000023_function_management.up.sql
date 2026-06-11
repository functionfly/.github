-- Create functions table
CREATE TABLE IF NOT EXISTS functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    providers TEXT[] NOT NULL DEFAULT '{}',
    region VARCHAR(100) NOT NULL DEFAULT 'global',
    code TEXT NOT NULL,
    env_vars JSONB DEFAULT '[]'::jsonb,
    version VARCHAR(50) DEFAULT '1.0.0',
    status VARCHAR(50) DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create function_deployments table
CREATE TABLE IF NOT EXISTS function_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES functions(id) ON DELETE CASCADE,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    provider VARCHAR(100) NOT NULL,
    region VARCHAR(100) NOT NULL,
    deployed_url TEXT,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create function_logs table
CREATE TABLE IF NOT EXISTS function_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID REFERENCES functions(id) ON DELETE CASCADE,
    deployment_id UUID REFERENCES function_deployments(id) ON DELETE CASCADE,
    level VARCHAR(20) NOT NULL DEFAULT 'info',
    message TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    source VARCHAR(100) DEFAULT 'system',
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_functions_tenant_id ON functions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_functions_status ON functions(status);
CREATE INDEX IF NOT EXISTS idx_function_deployments_function_id ON function_deployments(function_id);
CREATE INDEX IF NOT EXISTS idx_function_deployments_status ON function_deployments(status);
CREATE INDEX IF NOT EXISTS idx_function_logs_function_id ON function_logs(function_id);
CREATE INDEX IF NOT EXISTS idx_function_logs_deployment_id ON function_logs(deployment_id);
CREATE INDEX IF NOT EXISTS idx_function_logs_timestamp ON function_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_function_logs_level ON function_logs(level);

-- Create updated_at trigger function if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
DROP TRIGGER IF EXISTS update_functions_updated_at ON functions;
CREATE TRIGGER update_functions_updated_at BEFORE UPDATE ON functions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_function_deployments_updated_at ON function_deployments;
CREATE TRIGGER update_function_deployments_updated_at BEFORE UPDATE ON function_deployments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
