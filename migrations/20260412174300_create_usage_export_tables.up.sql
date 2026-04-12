-- Create usage_export_configurations table for billing data exports
CREATE TABLE IF NOT EXISTS usage_export_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    format VARCHAR(20) NOT NULL DEFAULT 'csv',
    data_types TEXT[] DEFAULT '{}',
    granularity VARCHAR(20),
    include_metadata BOOLEAN DEFAULT false,
    include_breakdown BOOLEAN DEFAULT false,
    date_range_type VARCHAR(30),
    function_filter UUID[] DEFAULT '{}',
    region_filter TEXT[] DEFAULT '{}',
    outcome_filter TEXT[] DEFAULT '{}',
    is_scheduled BOOLEAN DEFAULT false,
    schedule_frequency VARCHAR(20),
    schedule_day_of_month INTEGER,
    schedule_day_of_week INTEGER,
    schedule_hour INTEGER,
    delivery_method VARCHAR(30),
    email_recipients TEXT[] DEFAULT '{}',
    webhook_url TEXT,
    s3_bucket VARCHAR(255),
    s3_prefix VARCHAR(255),
    external_system_id UUID,
    field_mapping JSONB DEFAULT '{}',
    transform_config JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_executed_at TIMESTAMP WITH TIME ZONE,
    last_export_id UUID
);

CREATE TABLE IF NOT EXISTS usage_export_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    configuration_id UUID REFERENCES usage_export_configurations(id) ON DELETE SET NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    format VARCHAR(20) NOT NULL,
    data_types TEXT[] DEFAULT '{}',
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    record_count BIGINT DEFAULT 0,
    file_size_bytes BIGINT DEFAULT 0,
    storage_provider VARCHAR(30),
    storage_path TEXT,
    storage_url TEXT,
    checksum VARCHAR(64),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    notification_sent BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_export_configs_tenant ON usage_export_configurations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_export_configs_active ON usage_export_configurations(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_export_jobs_config ON usage_export_jobs(configuration_id);
CREATE INDEX IF NOT EXISTS idx_export_jobs_tenant ON usage_export_jobs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_export_jobs_status ON usage_export_jobs(status);
