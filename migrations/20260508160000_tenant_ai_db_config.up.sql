-- Platform-level migration: External AI database configuration
-- Stores encrypted connection details for tenant AI databases on Neon serverless PostgreSQL.
-- This table lives in the platform DB, NOT in tenant DBs.

CREATE TABLE IF NOT EXISTS tenant_ai_db_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE,
    provider VARCHAR(50) NOT NULL DEFAULT 'neon',  -- 'neon', 'turso', 'local'
    connection_details JSONB NOT NULL DEFAULT '{}', -- Encrypted connection string, project/branch IDs
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- 'pending', 'active', 'suspended', 'deleted'
    monthly_cost_estimate_cents INT DEFAULT 0,       -- Estimated monthly cost in cents
    markup_rate_bps INT DEFAULT 3000,                -- Markup rate in basis points (3000 = 30%)
    last_billed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_ai_db_config_tenant ON tenant_ai_db_config(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_ai_db_config_status ON tenant_ai_db_config(status);
