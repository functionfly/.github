-- Create external billing systems table
CREATE TABLE IF NOT EXISTS external_billing_systems (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    system_type VARCHAR(50) NOT NULL,
    api_endpoint TEXT,
    auth_type VARCHAR(30) NOT NULL,
    api_credential_key TEXT,
    api_credential_secret TEXT,
    oauth_token TEXT,
    oauth_refresh_token TEXT,
    oauth_expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    last_tested_at TIMESTAMP,
    last_test_status VARCHAR(30),
    last_test_error TEXT,
    sync_enabled BOOLEAN DEFAULT false,
    sync_frequency VARCHAR(20),
    sync_direction VARCHAR(20),
    last_sync_at TIMESTAMP,
    last_sync_status VARCHAR(30),
    field_mappings JSONB,
    value_mappings JSONB,
    transform_rules JSONB,
    webhook_secret TEXT,
    webhook_url TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    created_by UUID NOT NULL
);

-- Create billing integration syncs table
CREATE TABLE IF NOT EXISTS billing_integration_syncs (
    id UUID PRIMARY KEY,
    external_system_id UUID NOT NULL REFERENCES external_billing_systems(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    sync_type VARCHAR(30) NOT NULL,
    direction VARCHAR(20) NOT NULL,
    status VARCHAR(30) NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    records_processed BIGINT DEFAULT 0,
    records_created BIGINT DEFAULT 0,
    records_updated BIGINT DEFAULT 0,
    records_failed BIGINT DEFAULT 0,
    records_skipped BIGINT DEFAULT 0,
    error_message TEXT,
    error_details JSONB,
    external_batch_id VARCHAR(255),
    external_refs JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    triggered_by VARCHAR(30)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_external_billing_tenant ON external_billing_systems(tenant_id);
CREATE INDEX IF NOT EXISTS idx_billing_syncs_system ON billing_integration_syncs(external_system_id);
CREATE INDEX IF NOT EXISTS idx_billing_syncs_tenant ON billing_integration_syncs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_billing_syncs_status ON billing_integration_syncs(status);
