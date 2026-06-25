-- MicroVM usage tracking and billing tables for Enterprise tier
-- Tracks per-tenant MicroVM usage, execution records, and billing

BEGIN;

-- MicroVM execution records: individual function executions in MicroVMs
CREATE TABLE IF NOT EXISTS microvm_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_version VARCHAR(255) NOT NULL,
    execution_id UUID NOT NULL UNIQUE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    memory_mb INTEGER NOT NULL DEFAULT 512,
    vcpus INTEGER NOT NULL DEFAULT 2,
    status VARCHAR(50) NOT NULL DEFAULT 'running',
    outcome VARCHAR(50),
    error_message TEXT,
    network_allowed BOOLEAN NOT NULL DEFAULT false,
    packages_cached BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for MicroVM execution queries
CREATE INDEX IF NOT EXISTS idx_microvm_executions_tenant_id ON microvm_executions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_microvm_executions_function_id ON microvm_executions(function_id);
CREATE INDEX IF NOT EXISTS idx_microvm_executions_started_at ON microvm_executions(started_at);
CREATE INDEX IF NOT EXISTS idx_microvm_executions_tenant_started ON microvm_executions(tenant_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_microvm_executions_status ON microvm_executions(status);

-- MicroVM billing records: aggregated usage for billing cycle
CREATE TABLE IF NOT EXISTS microvm_billing_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    billing_period VARCHAR(7) NOT NULL,
    total_executions INTEGER NOT NULL DEFAULT 0,
    total_compute_seconds NUMERIC(12,2) NOT NULL DEFAULT 0,
    total_memory_seconds NUMERIC(12,2) NOT NULL DEFAULT 0,
    avg_memory_mb INTEGER NOT NULL DEFAULT 0,
    avg_vcpus NUMERIC(4,1) NOT NULL DEFAULT 0,
    base_fee_cents INTEGER NOT NULL DEFAULT 0,
    compute_charge_cents INTEGER NOT NULL DEFAULT 0,
    memory_charge_cents INTEGER NOT NULL DEFAULT 0,
    total_charge_cents INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, billing_period)
);

-- Indexes for MicroVM billing queries
CREATE INDEX IF NOT EXISTS idx_microvm_billing_tenant_period ON microvm_billing_records(tenant_id, billing_period DESC);
CREATE INDEX IF NOT EXISTS idx_microvm_billing_period ON microvm_billing_records(billing_period);

-- MicroVM audit log: security and compliance audit trail
CREATE TABLE IF NOT EXISTS microvm_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for MicroVM audit log
CREATE INDEX IF NOT EXISTS idx_microvm_audit_tenant_created ON microvm_audit_log(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_microvm_audit_action ON microvm_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_microvm_audit_resource ON microvm_audit_log(resource_type, resource_id);

-- MicroVM tenant quotas: per-tenant resource limits
CREATE TABLE IF NOT EXISTS microvm_tenant_quotas (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    max_concurrent_vms INTEGER NOT NULL DEFAULT 10,
    max_memory_mb INTEGER NOT NULL DEFAULT 2048,
    max_vcpus INTEGER NOT NULL DEFAULT 4,
    max_timeout_ms INTEGER NOT NULL DEFAULT 300000,
    current_compute_usage NUMERIC(12,2) NOT NULL DEFAULT 0,
    current_memory_usage NUMERIC(12,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for quota queries
CREATE INDEX IF NOT EXISTS idx_microvm_quotas_tenant ON microvm_tenant_quotas(tenant_id);

-- Function to update MicroVM usage after execution completes
CREATE OR REPLACE FUNCTION update_microvm_usage_after_execution()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND OLD.status = 'running' THEN
        -- Update tenant quotas with actual usage
        UPDATE microvm_tenant_quotas
        SET
            current_compute_usage = current_compute_usage + (NEW.duration_ms / 1000.0),
            current_memory_usage = current_memory_usage + ((NEW.duration_ms / 1000.0) * (NEW.memory_mb / 1024.0)),
            updated_at = NOW()
        WHERE tenant_id = NEW.tenant_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update usage on execution completion
DROP TRIGGER IF EXISTS trigger_update_microvm_usage ON microvm_executions;
CREATE TRIGGER trigger_update_microvm_usage
    AFTER UPDATE OF status ON microvm_executions
    FOR EACH ROW
    EXECUTE FUNCTION update_microvm_usage_after_execution();

-- Function to clean up old MicroVM executions (data retention)
CREATE OR REPLACE FUNCTION cleanup_microvm_executions(retention_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM microvm_executions
    WHERE started_at < NOW() - (retention_days || ' days')::INTERVAL
    AND status IN ('completed', 'failed', 'timeout');
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMIT;
