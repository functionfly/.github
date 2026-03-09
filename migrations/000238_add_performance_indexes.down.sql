-- Remove performance indexes (drop in reverse order of creation)

-- Dashboard config indexes
DROP INDEX IF EXISTS idx_dashboard_configs_type;
DROP INDEX IF EXISTS idx_dashboard_configs_tenant_user;

-- Function indexes
DROP INDEX IF EXISTS idx_function_logs_level_timestamp;
DROP INDEX IF EXISTS idx_function_logs_function_timestamp;
DROP INDEX IF EXISTS idx_function_deployments_created;
DROP INDEX IF EXISTS idx_function_deployments_function_status;
DROP INDEX IF EXISTS idx_functions_status;
DROP INDEX IF EXISTS idx_functions_tenant_app;

-- Local runtime indexes
DROP INDEX IF EXISTS idx_local_runtimes_last_heartbeat;
DROP INDEX IF EXISTS idx_local_runtimes_runtime_id;
DROP INDEX IF EXISTS idx_local_runtimes_status;

-- Session indexes
DROP INDEX IF EXISTS idx_sessions_expires;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_sessions_user_activity;

-- Security scan indexes
DROP INDEX IF EXISTS idx_vulnerabilities_category;
DROP INDEX IF EXISTS idx_vulnerabilities_scan_severity;
DROP INDEX IF EXISTS idx_security_scans_user_tenant;
DROP INDEX IF EXISTS idx_security_scans_status_started;

-- Monitoring indexes
DROP INDEX IF EXISTS idx_database_metrics_type_recorded;
DROP INDEX IF EXISTS idx_system_health_checks_status;
DROP INDEX IF EXISTS idx_system_health_checks_type_checked;

-- Feedback indexes
DROP INDEX IF EXISTS idx_feedback_type_priority;
DROP INDEX IF EXISTS idx_feedback_user_type;
DROP INDEX IF EXISTS idx_feedback_status_created;

-- Content management indexes
DROP INDEX IF EXISTS idx_blog_posts_tags;
DROP INDEX IF EXISTS idx_blog_posts_slug;
DROP INDEX IF EXISTS idx_blog_posts_published_at;
DROP INDEX IF EXISTS idx_changelog_entries_version;
DROP INDEX IF EXISTS idx_changelog_entries_published_date;

-- Alert indexes
DROP INDEX IF EXISTS idx_alerts_status_created;
DROP INDEX IF EXISTS idx_alerts_type_severity;
DROP INDEX IF EXISTS idx_alerts_tenant_status;

-- Performance metric indexes
DROP INDEX IF EXISTS idx_performance_metrics_app_timestamp;
DROP INDEX IF EXISTS idx_performance_metrics_tenant_timestamp;
DROP INDEX IF EXISTS idx_performance_metrics_type_timestamp;

-- Usage event indexes
DROP INDEX IF EXISTS idx_usage_rollups_tenant_period;
DROP INDEX IF EXISTS idx_usage_events_tenant_timestamp;
DROP INDEX IF EXISTS idx_usage_events_tenant_type_timestamp;

-- Billing indexes
DROP INDEX IF EXISTS idx_invoices_due_date;
DROP INDEX IF EXISTS idx_invoices_tenant_status;
DROP INDEX IF EXISTS idx_subscriptions_period;
DROP INDEX IF EXISTS idx_subscriptions_tenant_status;

-- Audit event indexes
DROP INDEX IF EXISTS idx_audit_events_actor_tenant;
DROP INDEX IF EXISTS idx_audit_events_action_timestamp;
DROP INDEX IF EXISTS idx_audit_events_resource;
DROP INDEX IF EXISTS idx_audit_events_tenant_timestamp;

-- Deployment indexes
DROP INDEX IF EXISTS idx_deployments_provider_region;
DROP INDEX IF EXISTS idx_deployments_app_created;
DROP INDEX IF EXISTS idx_deployments_app_status;

-- Routing event indexes
DROP INDEX IF EXISTS idx_routing_events_timestamp_outcome;
DROP INDEX IF EXISTS idx_routing_events_backend_timestamp;
DROP INDEX IF EXISTS idx_routing_events_app_timestamp;

-- Circuit state indexes
DROP INDEX IF EXISTS idx_circuit_states_state;
DROP INDEX IF EXISTS idx_circuit_states_backend;

-- Health check indexes
DROP INDEX IF EXISTS idx_health_checks_timestamp_ok;
DROP INDEX IF EXISTS idx_health_checks_backend_timestamp;

-- Backend indexes
DROP INDEX IF EXISTS idx_backends_enabled_priority;
DROP INDEX IF EXISTS idx_backends_app_enabled;

-- App indexes
DROP INDEX IF EXISTS idx_apps_tenant_name;
DROP INDEX IF EXISTS idx_apps_tenant_slug;

-- Tenant indexes
DROP INDEX IF EXISTS idx_tenants_status;

-- User indexes
DROP INDEX IF EXISTS idx_users_verification_token;
DROP INDEX IF EXISTS idx_users_provider;
DROP INDEX IF EXISTS idx_users_email_verified;
DROP INDEX IF EXISTS idx_users_tenant_email;

-- Partial indexes
DROP INDEX IF EXISTS idx_performance_metrics_recent;
DROP INDEX IF EXISTS idx_alerts_active;
DROP INDEX IF EXISTS idx_backends_active;
DROP INDEX IF EXISTS idx_deployments_active;