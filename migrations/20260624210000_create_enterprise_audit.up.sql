-- Create enterprise audit log table for comprehensive tenant activity tracking
CREATE TABLE IF NOT EXISTS enterprise_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    -- Service area (auth, vault, billing, functions, etc.)
    service_area VARCHAR(50) NOT NULL,

    -- Action performed (create, read, update, delete, etc.)
    action VARCHAR(100) NOT NULL,

    -- Resource type (user, team, function, secret, etc.)
    resource_type VARCHAR(50) NOT NULL,

    -- Resource ID (optional - null for some actions)
    resource_id UUID,

    -- Actor information
    actor_type VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    actor_name VARCHAR(255),

    -- Request context
    request_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Additional metadata (JSONB for flexibility)
    metadata JSONB DEFAULT '{}',

    -- Success/failure
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,

    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Foreign key to tenant
    CONSTRAINT fk_enterprise_audit_log_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_enterprise_audit_log_tenant_id ON enterprise_audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_enterprise_audit_log_tenant_created ON enterprise_audit_log(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_enterprise_audit_log_service_area ON enterprise_audit_log(tenant_id, service_area, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_enterprise_audit_log_action ON enterprise_audit_log(tenant_id, action, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_enterprise_audit_log_resource ON enterprise_audit_log(tenant_id, resource_type, resource_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_enterprise_audit_log_actor ON enterprise_audit_log(tenant_id, actor_type, actor_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_enterprise_audit_log_created_at ON enterprise_audit_log(created_at);

-- Comment on table
COMMENT ON TABLE enterprise_audit_log IS 'Enterprise audit log for comprehensive tenant activity tracking';
COMMENT ON COLUMN enterprise_audit_log.service_area IS 'Service area: auth, vault, billing, functions, registry, agents, teams, api, sso, scim, webhook, settings, system';
COMMENT ON COLUMN enterprise_audit_log.action IS 'Action performed: create, read, update, delete, login, logout, export, etc.';
COMMENT ON COLUMN enterprise_audit_log.resource_type IS 'Type of resource: user, team, function, secret, api_key, app, etc.';
