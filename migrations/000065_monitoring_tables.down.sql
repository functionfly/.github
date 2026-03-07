-- Rollback monitoring tables migration

-- Drop realtime publications (if configured)
-- Note: These would be removed in Supabase dashboard or via separate migration

-- Drop indexes
DROP INDEX IF EXISTS idx_performance_metrics_type_timestamp;
DROP INDEX IF EXISTS idx_performance_metrics_tenant;
DROP INDEX IF EXISTS idx_performance_metrics_app;
DROP INDEX IF EXISTS idx_performance_metrics_backend;
DROP INDEX IF EXISTS idx_alerts_type_status;
DROP INDEX IF EXISTS idx_alerts_severity;
DROP INDEX IF EXISTS idx_alerts_tenant;
DROP INDEX IF EXISTS idx_alerts_created;
DROP INDEX IF EXISTS idx_system_health_checks_type;
DROP INDEX IF EXISTS idx_system_health_checks_status;
DROP INDEX IF EXISTS idx_system_health_checks_checked;
DROP INDEX IF EXISTS idx_monitoring_events_type;
DROP INDEX IF EXISTS idx_monitoring_events_tenant;
DROP INDEX IF EXISTS idx_monitoring_events_timestamp;
DROP INDEX IF EXISTS idx_dashboard_configs_tenant;
DROP INDEX IF EXISTS idx_dashboard_configs_user;

-- Drop tables
DROP TABLE IF EXISTS dashboard_configs;
DROP TABLE IF EXISTS monitoring_events;
DROP TABLE IF EXISTS system_health_checks;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS performance_metrics;