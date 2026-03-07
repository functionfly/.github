-- Create SCIM configuration table
CREATE TABLE scim_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    enabled BOOLEAN DEFAULT false,
    idp_url VARCHAR(500),
    idp_token VARCHAR(500),
    secret_key BYTEA,
    sync_groups BOOLEAN DEFAULT true,
    sync_users BOOLEAN DEFAULT true,
    last_sync_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create SCIM sync log table
CREATE TABLE scim_sync_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    direction VARCHAR(20),
    resource_type VARCHAR(50),
    resource_id VARCHAR(255),
    action VARCHAR(20),
    success BOOLEAN,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for SCIM configs
CREATE INDEX IF NOT EXISTS idx_scim_configs_tenant_id ON scim_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scim_configs_enabled ON scim_configs(enabled) WHERE enabled = true;

-- Create indexes for SCIM sync logs
CREATE INDEX IF NOT EXISTS idx_scim_sync_log_tenant_id ON scim_sync_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scim_sync_log_created_at ON scim_sync_log(created_at);
CREATE INDEX IF NOT EXISTS idx_scim_sync_log_direction ON scim_sync_log(direction);
CREATE INDEX IF NOT EXISTS idx_scim_sync_log_resource ON scim_sync_log(resource_type, resource_id);

-- Add unique constraint for tenant SCIM config (one config per tenant)
ALTER TABLE scim_configs ADD CONSTRAINT uq_scim_configs_tenant UNIQUE (tenant_id);
