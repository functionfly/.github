-- Rollback SIEM configuration and export log tables
-- Migration: 000085_create_siem_configs

-- Drop RLS policies first
DROP POLICY IF EXISTS siem_configs_tenant_policy ON siem_configs;
DROP POLICY IF EXISTS siem_export_logs_tenant_policy ON siem_export_logs;

-- Disable row level security
ALTER TABLE siem_configs DISABLE ROW LEVEL SECURITY;
ALTER TABLE siem_export_logs DISABLE ROW LEVEL SECURITY;

-- Drop indexes
DROP INDEX IF EXISTS idx_siem_configs_tenant_id;
DROP INDEX IF EXISTS idx_siem_configs_enabled;
DROP INDEX IF EXISTS idx_siem_configs_destination_type;
DROP INDEX IF EXISTS idx_siem_export_logs_siem_config_id;
DROP INDEX IF EXISTS idx_siem_export_logs_status;
DROP INDEX IF EXISTS idx_siem_export_logs_created_at;

-- Drop tables
DROP TABLE IF EXISTS siem_export_logs;
DROP TABLE IF EXISTS siem_configs;
