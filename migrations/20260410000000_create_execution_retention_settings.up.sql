-- Execution Log Retention Settings Table
-- Stores configurable retention policies for execution logs and related data

CREATE TABLE IF NOT EXISTS execution_retention_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Table-specific retention periods (in days, NULL or 0 = no cleanup)
    execution_retention_days INT NOT NULL DEFAULT 90,
    public_execution_retention_days INT NOT NULL DEFAULT 30,
    resource_usage_retention_days INT NOT NULL DEFAULT 90,
    meg_record_retention_days INT NOT NULL DEFAULT 365,
    drift_report_retention_days INT NOT NULL DEFAULT 365,
    execution_cert_retention_days INT NOT NULL DEFAULT 365,

    -- Cleanup settings
    cleanup_interval_minutes INT NOT NULL DEFAULT 1440,  -- 24 hours
    batch_size INT NOT NULL DEFAULT 1000,
    verbose_logging BOOLEAN NOT NULL DEFAULT FALSE,

    -- Lifecycle tracking
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Single-row constraint (only one global settings row)
    is_active BOOLEAN NOT NULL DEFAULT TRUE UNIQUE
);

-- Ensure only one active settings row exists using a partial unique index
CREATE UNIQUE INDEX IF NOT EXISTS idx_execution_retention_settings_single_active
ON execution_retention_settings (is_active)
WHERE is_active = TRUE;

-- Add timestamp index for audit queries
CREATE INDEX IF NOT EXISTS idx_execution_retention_settings_updated_at
ON execution_retention_settings (updated_at DESC);

-- Insert default settings row (single row enforced by unique constraint)
INSERT INTO execution_retention_settings (
    execution_retention_days,
    public_execution_retention_days,
    resource_usage_retention_days,
    meg_record_retention_days,
    drift_report_retention_days,
    execution_cert_retention_days,
    cleanup_interval_minutes,
    batch_size,
    verbose_logging,
    is_active
) VALUES (
    90,   -- execution_retention_days
    30,   -- public_execution_retention_days
    90,   -- resource_usage_retention_days
    365,  -- meg_record_retention_days
    365,  -- drift_report_retention_days
    365,  -- execution_cert_retention_days
    1440, -- cleanup_interval_minutes (24 hours)
    1000, -- batch_size
    FALSE, -- verbose_logging
    TRUE  -- is_active (enforces single row)
)
ON CONFLICT (is_active) DO NOTHING;

COMMENT ON TABLE execution_retention_settings IS 'Global configuration for execution log retention policies';
COMMENT ON COLUMN execution_retention_settings.execution_retention_days IS 'Days to keep records in registry_function_executions (0 = disable cleanup)';
COMMENT ON COLUMN execution_retention_settings.public_execution_retention_days IS 'Days to keep records in registry_executions_public (0 = disable cleanup)';
COMMENT ON COLUMN execution_retention_settings.resource_usage_retention_days IS 'Days to keep records in execution_resource_usage (0 = disable cleanup)';
COMMENT ON COLUMN execution_retention_settings.meg_record_retention_days IS 'Days to keep records in execution_meg_records (0 = disable cleanup)';
COMMENT ON COLUMN execution_retention_settings.drift_report_retention_days IS 'Days to keep records in drift_reports (0 = disable cleanup)';
COMMENT ON COLUMN execution_retention_settings.execution_cert_retention_days IS 'Days to keep records in execution_certificates (0 = disable cleanup)';
COMMENT ON COLUMN execution_retention_settings.cleanup_interval_minutes IS 'Interval between cleanup runs in minutes';
COMMENT ON COLUMN execution_retention_settings.batch_size IS 'Number of records to delete per batch';
COMMENT ON COLUMN execution_retention_settings.verbose_logging IS 'Enable detailed logging of cleanup operations';
