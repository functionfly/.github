-- Create SIEM configuration and export log tables
-- Migration: 000085_create_siem_configs
-- Uses current_tenant_id() for RLS (session app.tenant_id), not auth.users

-- SIEM Configuration table
CREATE TABLE IF NOT EXISTS siem_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT false,
    export_format VARCHAR(20) DEFAULT 'json',
    destination_type VARCHAR(50) NOT NULL,
    config JSONB DEFAULT '{}',
    last_export_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- SIEM Export Log table
CREATE TABLE IF NOT EXISTS siem_export_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    siem_config_id UUID NOT NULL REFERENCES siem_configs(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    records_sent INT DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for SIEM configs
CREATE INDEX IF NOT EXISTS idx_siem_configs_tenant_id ON siem_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_siem_configs_enabled ON siem_configs(enabled);
CREATE INDEX IF NOT EXISTS idx_siem_configs_destination_type ON siem_configs(destination_type);

-- Indexes for SIEM export logs
CREATE INDEX IF NOT EXISTS idx_siem_export_logs_siem_config_id ON siem_export_logs(siem_config_id);
CREATE INDEX IF NOT EXISTS idx_siem_export_logs_status ON siem_export_logs(status);
CREATE INDEX IF NOT EXISTS idx_siem_export_logs_created_at ON siem_export_logs(created_at);

-- Row level security for tenant isolation
ALTER TABLE siem_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE siem_export_logs ENABLE ROW LEVEL SECURITY;

-- Policy for SIEM configs - tenant access (uses session tenant from app.tenant_id)
DROP POLICY IF EXISTS siem_configs_tenant_policy ON siem_configs;
CREATE POLICY siem_configs_tenant_policy ON siem_configs
    FOR ALL
    USING (tenant_id = current_tenant_id());

-- Policy for SIEM export logs - tenant access via config
DROP POLICY IF EXISTS siem_export_logs_tenant_policy ON siem_export_logs;
CREATE POLICY siem_export_logs_tenant_policy ON siem_export_logs
    FOR ALL
    USING (
        siem_config_id IN (
            SELECT id FROM siem_configs WHERE tenant_id = current_tenant_id()
        )
    );
