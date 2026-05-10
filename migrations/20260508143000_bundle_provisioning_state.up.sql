-- Platform-level migration: Bundle provisioning state tracking
-- This table lives in the platform DB and tracks the provisioning status
-- of each tenant's isolated bundle components.

CREATE TABLE IF NOT EXISTS tenant_bundle_state (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE,
    bundle_slug VARCHAR(100) NOT NULL,
    provision_status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'provisioning', 'active', 'failed', 'suspended'
    provisioned_at TIMESTAMP WITH TIME ZONE,
    components JSONB NOT NULL DEFAULT '{}', -- {auth: {status, timestamp}, payments: {...}, ...}
    error_log JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_bundle_state_tenant ON tenant_bundle_state(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_bundle_state_status ON tenant_bundle_state(provision_status);
CREATE INDEX IF NOT EXISTS idx_tenant_bundle_state_slug ON tenant_bundle_state(bundle_slug);
