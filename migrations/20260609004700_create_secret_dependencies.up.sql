-- Create secret_dependencies table for tracking which services depend on secrets
-- This enables impact analysis during rotation/deletion and helps prevent outages

CREATE TABLE IF NOT EXISTS secret_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets_vault(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    dependent_id UUID NOT NULL,
    dependent_type VARCHAR(50) NOT NULL, -- 'function', 'service', 'integration', 'workflow'
    dependent_name VARCHAR(255) NOT NULL,
    criticality VARCHAR(20) NOT NULL DEFAULT 'medium', -- 'low', 'medium', 'high', 'critical'
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_secret FOREIGN KEY (secret_id) REFERENCES secrets_vault(id) ON DELETE CASCADE
);

CREATE INDEX idx_secret_dependencies_secret_id ON secret_dependencies(secret_id);
CREATE INDEX idx_secret_dependencies_tenant_id ON secret_dependencies(tenant_id);
CREATE INDEX idx_secret_dependencies_dependent_id ON secret_dependencies(dependent_id);
CREATE INDEX idx_secret_dependencies_dependent_type ON secret_dependencies(dependent_type);
CREATE UNIQUE INDEX idx_secret_dependencies_unique ON secret_dependencies(secret_id, dependent_id, dependent_type);

COMMENT ON TABLE secret_dependencies IS 'Tracks which services/functions depend on which secrets for impact analysis during rotation or deletion';