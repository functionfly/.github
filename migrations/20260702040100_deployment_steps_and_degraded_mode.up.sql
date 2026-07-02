-- Deployment steps tracking for real-time provisioning status
CREATE TABLE IF NOT EXISTS deployment_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    deployment_id UUID NOT NULL,
    bundle_slug VARCHAR(100) NOT NULL,
    step_name VARCHAR(100) NOT NULL,
    step_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE deployment_steps ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;
ALTER TABLE deployment_steps ADD COLUMN IF NOT EXISTS error_message TEXT;

CREATE INDEX IF NOT EXISTS idx_deployment_steps_tenant_id ON deployment_steps(tenant_id);
CREATE INDEX IF NOT EXISTS idx_deployment_steps_deployment_id ON deployment_steps(deployment_id);
CREATE INDEX IF NOT EXISTS idx_deployment_steps_status ON deployment_steps(status);

-- Add degraded_mode tracking to tenants
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS degraded_mode BOOLEAN DEFAULT false;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS degradation_reason TEXT;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS degradation_updated_at TIMESTAMPTZ;
