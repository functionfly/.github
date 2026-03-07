-- Add deployment tracking tables

-- Deployments table
CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id UUID NOT NULL REFERENCES apps(id),
    provider VARCHAR(50) NOT NULL, -- 'workers', 'vercel', 'fly'
    region VARCHAR(50) NOT NULL,
    deployment_id VARCHAR(255) NOT NULL, -- Provider-specific deployment ID
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'deploying', 'success', 'failed', 'rollback'
    artifact_key VARCHAR(500), -- Reference to stored artifact
    routes JSONB, -- Route patterns as JSON array
    message TEXT, -- Status message or error details
    metadata JSONB, -- Additional metadata from provider
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Deployment artifacts table
CREATE TABLE IF NOT EXISTS deployment_artifacts (
    key VARCHAR(500) PRIMARY KEY, -- Artifact storage key
    app_id UUID NOT NULL REFERENCES apps(id),
    provider VARCHAR(50) NOT NULL,
    content_type VARCHAR(100),
    size BIGINT,
    checksum VARCHAR(255), -- SHA-256 or similar
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_deployments_app_id ON deployments(app_id);
CREATE INDEX IF NOT EXISTS idx_deployments_provider ON deployments(provider);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
CREATE INDEX IF NOT EXISTS idx_deployments_created_at ON deployments(created_at);
CREATE INDEX IF NOT EXISTS idx_deployment_artifacts_app_id ON deployment_artifacts(app_id);