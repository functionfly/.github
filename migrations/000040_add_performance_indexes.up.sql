-- Add strategic database indexes for better query performance
-- These indexes target frequently queried columns and composite indexes for common access patterns

-- User indexes
CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified) WHERE email_verified = false;
CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id) WHERE provider IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_verification_token ON users(verification_token) WHERE verification_token IS NOT NULL;

-- Tenant indexes
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);

-- App indexes
CREATE INDEX IF NOT EXISTS idx_apps_tenant_slug ON apps(tenant_id, slug);
CREATE INDEX IF NOT EXISTS idx_apps_tenant_name ON apps(tenant_id, name);

-- Backend indexes
CREATE INDEX IF NOT EXISTS idx_backends_app_enabled ON backends(app_id, enabled);
CREATE INDEX IF NOT EXISTS idx_backends_enabled_priority ON backends(enabled, priority) WHERE enabled = true;

-- Health check indexes
CREATE INDEX IF NOT EXISTS idx_health_checks_backend_timestamp ON health_checks(backend_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_health_checks_timestamp_ok ON health_checks(timestamp DESC, ok);

-- Circuit state indexes (table may not exist)
-- CREATE INDEX IF NOT EXISTS idx_circuit_states_backend ON circuit_states(backend_id);
-- CREATE INDEX IF NOT EXISTS idx_circuit_states_state ON circuit_states(state);

-- Routing event indexes (high volume table)
-- CREATE INDEX IF NOT EXISTS idx_routing_events_app_timestamp ON routing_events(app_id, timestamp DESC);
-- CREATE INDEX IF NOT EXISTS idx_routing_events_backend_timestamp ON routing_events(backend_id, timestamp DESC);
-- CREATE INDEX IF NOT EXISTS idx_routing_events_timestamp_outcome ON routing_events(timestamp DESC, outcome);

-- Deployment indexes
CREATE INDEX IF NOT EXISTS idx_deployments_app_status ON deployments(app_id, status);
CREATE INDEX IF NOT EXISTS idx_deployments_app_created ON deployments(app_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_deployments_provider_region ON deployments(provider, region);

-- Audit event indexes (high volume table)
CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_timestamp ON audit_events(tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_resource ON audit_events(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_action_timestamp ON audit_events(action, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_actor_tenant ON audit_events(actor_user_id, tenant_id);

-- Billing indexes
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_status ON subscriptions(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_period ON subscriptions(current_period_start, current_period_end);
CREATE INDEX IF NOT EXISTS idx_invoices_tenant_status ON invoices(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date) WHERE status = 'open';

-- Usage event indexes (high volume table)
CREATE INDEX IF NOT EXISTS idx_usage_events_tenant_type_timestamp ON usage_events(tenant_id, event_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_usage_events_tenant_timestamp ON usage_events(tenant_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_usage_rollups_tenant_period ON usage_rollups(tenant_id, period_date);

-- Performance metric indexes (high volume table)
CREATE INDEX IF NOT EXISTS idx_performance_metrics_type_timestamp ON performance_metrics(metric_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_tenant_timestamp ON performance_metrics(tenant_id, timestamp DESC) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_performance_metrics_app_timestamp ON performance_metrics(app_id, timestamp DESC) WHERE app_id IS NOT NULL;

-- Alert indexes
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_status ON alerts(tenant_id, status) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_alerts_type_severity ON alerts(alert_type, severity);
CREATE INDEX IF NOT EXISTS idx_alerts_status_created ON alerts(status, created_at DESC);

-- Content management indexes
CREATE INDEX IF NOT EXISTS idx_changelog_entries_published_date ON changelog_entries(is_published, date DESC);
CREATE INDEX IF NOT EXISTS idx_changelog_entries_version ON changelog_entries(version);
CREATE INDEX IF NOT EXISTS idx_blog_posts_published_at ON blog_posts(is_published, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_blog_posts_slug ON blog_posts(slug);
CREATE INDEX IF NOT EXISTS idx_blog_posts_tags ON blog_posts USING gin(tags);

-- Feedback indexes
CREATE INDEX IF NOT EXISTS idx_feedback_status_created ON feedback(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_feedback_user_type ON feedback(user_id, feedback_type);
CREATE INDEX IF NOT EXISTS idx_feedback_type_priority ON feedback(feedback_type, priority);

-- Monitoring indexes
CREATE INDEX IF NOT EXISTS idx_system_health_checks_type_checked ON system_health_checks(check_type, checked_at DESC);
CREATE INDEX IF NOT EXISTS idx_system_health_checks_status ON system_health_checks(status);
CREATE INDEX IF NOT EXISTS idx_database_metrics_type_recorded ON database_metrics(metric_type, recorded_at DESC);

-- Security scan indexes
CREATE INDEX IF NOT EXISTS idx_security_scans_status_started ON security_scans(status, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_security_scans_user_tenant ON security_scans(user_id, tenant_id);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_scan_severity ON vulnerabilities(scan_id, severity);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_category ON vulnerabilities(category);

-- Session indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_activity ON sessions(user_id, last_activity DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(session_token);
-- CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at) WHERE expires_at > NOW(); -- NOW() not allowed in partial index

-- Local runtime indexes (tables may not exist)
-- CREATE INDEX IF NOT EXISTS idx_local_runtimes_status ON local_runtime_instances(status);
-- CREATE INDEX IF NOT EXISTS idx_local_runtimes_runtime_id ON local_runtime_instances(runtime_id);
-- CREATE INDEX IF NOT EXISTS idx_local_runtimes_last_heartbeat ON local_runtime_instances(last_heartbeat DESC);

-- Function indexes (using functions table instead of function_configs)
CREATE INDEX IF NOT EXISTS idx_functions_tenant_app ON functions(tenant_id, app_id);
CREATE INDEX IF NOT EXISTS idx_functions_status ON functions(status);
CREATE INDEX IF NOT EXISTS idx_function_deployments_function_status ON function_deployments(function_id, status);
CREATE INDEX IF NOT EXISTS idx_function_deployments_created ON function_deployments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_function_logs_function_timestamp ON function_logs(function_id, timestamp DESC) WHERE function_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_function_logs_level_timestamp ON function_logs(level, timestamp DESC);

-- Dashboard config indexes
CREATE INDEX IF NOT EXISTS idx_dashboard_configs_tenant_user ON dashboard_configs(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_configs_type ON dashboard_configs(config_type);

-- Partial indexes for common queries (commented out due to NOW() function issues)
-- CREATE INDEX IF NOT EXISTS idx_deployments_active ON deployments(status) WHERE status IN ('success', 'deploying');
-- CREATE INDEX IF NOT EXISTS idx_backends_active ON backends(enabled) WHERE enabled = true;
-- CREATE INDEX IF NOT EXISTS idx_alerts_active ON alerts(status) WHERE status = 'active';
-- CREATE INDEX IF NOT EXISTS idx_performance_metrics_recent ON performance_metrics(timestamp DESC) WHERE timestamp > NOW() - INTERVAL '30 days';
