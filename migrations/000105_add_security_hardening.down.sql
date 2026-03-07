-- Remove Row Level Security and audit triggers

-- Drop security monitoring view
DROP VIEW IF EXISTS security_monitoring;

-- Drop security functions
DROP FUNCTION IF EXISTS get_tenant_security_summary(uuid);
DROP FUNCTION IF EXISTS cleanup_expired_sessions();
DROP FUNCTION IF EXISTS validate_function_security(text);
DROP FUNCTION IF EXISTS check_tenant_rate_limit(text, integer, interval);
DROP FUNCTION IF EXISTS validate_tenant_access(uuid);
DROP FUNCTION IF EXISTS audit_trigger_function();
DROP FUNCTION IF EXISTS is_platform_admin();
DROP FUNCTION IF EXISTS current_user_id();
DROP FUNCTION IF EXISTS current_tenant_id();

-- Drop audit triggers
DROP TRIGGER IF EXISTS audit_users_trigger ON users;
DROP TRIGGER IF EXISTS audit_apps_trigger ON apps;
DROP TRIGGER IF EXISTS audit_backends_trigger ON backends;
DROP TRIGGER IF EXISTS audit_deployments_trigger ON deployments;
DROP TRIGGER IF EXISTS audit_subscriptions_trigger ON subscriptions;
DROP TRIGGER IF EXISTS audit_function_configs_trigger ON function_configs;
DROP TRIGGER IF EXISTS audit_function_deployments_trigger ON function_deployments;

-- Drop Row Level Security policies
DROP POLICY IF EXISTS users_tenant_isolation ON users;
DROP POLICY IF EXISTS users_self_access ON users;
DROP POLICY IF EXISTS apps_tenant_isolation ON apps;
DROP POLICY IF EXISTS backends_tenant_isolation ON backends;
DROP POLICY IF EXISTS deployments_tenant_isolation ON deployments;
DROP POLICY IF EXISTS audit_events_tenant_isolation ON audit_events;
DROP POLICY IF EXISTS usage_events_tenant_isolation ON usage_events;
DROP POLICY IF EXISTS subscriptions_tenant_isolation ON subscriptions;
DROP POLICY IF EXISTS invoices_tenant_isolation ON invoices;
DROP POLICY IF EXISTS feedback_tenant_isolation ON feedback;
DROP POLICY IF EXISTS changelog_entries_public_read ON changelog_entries;
DROP POLICY IF EXISTS blog_posts_public_read ON blog_posts;
DROP POLICY IF EXISTS performance_metrics_tenant_isolation ON performance_metrics;
DROP POLICY IF EXISTS alerts_tenant_isolation ON alerts;
DROP POLICY IF EXISTS security_scans_tenant_isolation ON security_scans;
DROP POLICY IF EXISTS vulnerabilities_tenant_isolation ON vulnerabilities;
DROP POLICY IF EXISTS dashboard_configs_tenant_isolation ON dashboard_configs;
DROP POLICY IF EXISTS function_configs_tenant_isolation ON function_configs;
DROP POLICY IF EXISTS function_deployments_tenant_isolation ON function_deployments;
DROP POLICY IF EXISTS function_logs_tenant_isolation ON function_logs;

-- Disable Row Level Security on tables
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE apps DISABLE ROW LEVEL SECURITY;
ALTER TABLE backends DISABLE ROW LEVEL SECURITY;
ALTER TABLE deployments DISABLE ROW LEVEL SECURITY;
ALTER TABLE audit_events DISABLE ROW LEVEL SECURITY;
ALTER TABLE usage_events DISABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions DISABLE ROW LEVEL SECURITY;
ALTER TABLE invoices DISABLE ROW LEVEL SECURITY;
ALTER TABLE feedback DISABLE ROW LEVEL SECURITY;
ALTER TABLE changelog_entries DISABLE ROW LEVEL SECURITY;
ALTER TABLE blog_posts DISABLE ROW LEVEL SECURITY;
ALTER TABLE performance_metrics DISABLE ROW LEVEL SECURITY;
ALTER TABLE alerts DISABLE ROW LEVEL SECURITY;
ALTER TABLE security_scans DISABLE ROW LEVEL SECURITY;
ALTER TABLE vulnerabilities DISABLE ROW LEVEL SECURITY;
ALTER TABLE dashboard_configs DISABLE ROW LEVEL SECURITY;
ALTER TABLE function_configs DISABLE ROW LEVEL SECURITY;
ALTER TABLE function_deployments DISABLE ROW LEVEL SECURITY;
ALTER TABLE function_logs DISABLE ROW LEVEL SECURITY;

-- Remove security check constraint
ALTER TABLE function_configs DROP CONSTRAINT IF EXISTS function_code_security_check;

-- Remove RLS support indexes (keep performance indexes)
DROP INDEX IF EXISTS idx_users_tenant_id;
DROP INDEX IF EXISTS idx_apps_tenant_id;
DROP INDEX IF EXISTS idx_backends_app_id;
DROP INDEX IF EXISTS idx_deployments_app_id;
DROP INDEX IF EXISTS idx_subscriptions_tenant_id;
DROP INDEX IF EXISTS idx_invoices_tenant_id;
DROP INDEX IF EXISTS idx_dashboard_configs_tenant_id;
DROP INDEX IF EXISTS idx_function_configs_tenant_id;
DROP INDEX IF EXISTS idx_function_deployments_function_id;