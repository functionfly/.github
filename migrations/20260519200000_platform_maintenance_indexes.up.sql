-- Platform Maintenance Indexes
-- Additional indexes for platform maintenance queries

CREATE INDEX IF NOT EXISTS idx_platform_maintenance_enabled ON platform_maintenance(enabled) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_platform_maintenance_scheduled ON platform_maintenance(scheduled_start) WHERE is_scheduled = true;
CREATE INDEX IF NOT EXISTS idx_maintenance_audit_log_maintenance_id ON maintenance_audit_log(maintenance_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_audit_log_changed_at ON maintenance_audit_log(changed_at DESC);
