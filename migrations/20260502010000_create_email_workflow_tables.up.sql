-- Create email workflow configs table for tenant-isolated email workflows
-- Each bundle type has pre-configured workflows that are auto-provisioned on deploy

CREATE TABLE IF NOT EXISTS email_workflow_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    bundle_slug VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    trigger VARCHAR(50) NOT NULL, -- 'on_signup', 'on_payment', 'on_milestone', 'on_inactivity', 'manual'
    category VARCHAR(50) NOT NULL, -- 'onboarding', 'billing', 'engagement', 'retention', 'security'
    delay_days INTEGER DEFAULT 0, -- 0 = immediate, positive = delay in days
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_workflow_configs_tenant_id ON email_workflow_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_email_workflow_configs_bundle_slug ON email_workflow_configs(bundle_slug);
CREATE INDEX IF NOT EXISTS idx_email_workflow_configs_active ON email_workflow_configs(active);

-- Create email workflow executions table for tracking workflow runs
CREATE TABLE IF NOT EXISTS email_workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    workflow_id UUID NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'sent', 'failed', 'cancelled'
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    retry_count INTEGER DEFAULT 0,
    last_retry_at TIMESTAMP WITH TIME ZONE,
    email_subject VARCHAR(500),
    email_template VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_workflow_executions_tenant_id ON email_workflow_executions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_email_workflow_executions_workflow_id ON email_workflow_executions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_email_workflow_executions_status ON email_workflow_executions(status);
CREATE INDEX IF NOT EXISTS idx_email_workflow_executions_scheduled_at ON email_workflow_executions(scheduled_at);
