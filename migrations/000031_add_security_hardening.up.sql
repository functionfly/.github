-- Add Row Level Security and audit triggers for multi-tenant data protection

-- Enable Row Level Security on multi-tenant tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE apps ENABLE ROW LEVEL SECURITY;
ALTER TABLE backends ENABLE ROW LEVEL SECURITY;
ALTER TABLE deployments ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE feedback ENABLE ROW LEVEL SECURITY;
ALTER TABLE changelog_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE blog_posts ENABLE ROW LEVEL SECURITY;
ALTER TABLE performance_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_scans ENABLE ROW LEVEL SECURITY;
ALTER TABLE vulnerabilities ENABLE ROW LEVEL SECURITY;
ALTER TABLE dashboard_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE function_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE function_deployments ENABLE ROW LEVEL SECURITY;
ALTER TABLE function_logs ENABLE ROW LEVEL SECURITY;

-- Create a function to get current tenant ID from session
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS uuid AS $$
BEGIN
    -- In production, this should be set by your application middleware
    -- For now, we'll use a session variable that can be set by the application
    RETURN NULLIF(current_setting('app.current_tenant_id', true), '')::uuid;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create a function to get current user ID from session
CREATE OR REPLACE FUNCTION current_user_id()
RETURNS uuid AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_user_id', true), '')::uuid;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create a function to check if current user is a platform admin
CREATE OR REPLACE FUNCTION is_platform_admin()
RETURNS boolean AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM users
        WHERE id = current_user_id()
        AND role IN ('admin', 'super_admin')
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Row Level Security Policies for Users table
CREATE POLICY users_tenant_isolation ON users
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR is_platform_admin()
    );

CREATE POLICY users_self_access ON users
    FOR ALL USING (id = current_user_id());

-- Row Level Security Policies for Apps table
CREATE POLICY apps_tenant_isolation ON apps
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR is_platform_admin()
    );

-- Row Level Security Policies for Backends table
CREATE POLICY backends_tenant_isolation ON backends
    FOR ALL USING (
        app_id IN (
            SELECT id FROM apps WHERE tenant_id = current_tenant_id()
        )
        OR is_platform_admin()
    );

-- Row Level Security Policies for Deployments table
CREATE POLICY deployments_tenant_isolation ON deployments
    FOR ALL USING (
        app_id IN (
            SELECT id FROM apps WHERE tenant_id = current_tenant_id()
        )
        OR is_platform_admin()
    );

-- Row Level Security Policies for Audit Events table
CREATE POLICY audit_events_tenant_isolation ON audit_events
    FOR SELECT USING (
        tenant_id = current_tenant_id()
        OR is_platform_admin()
    );

-- Row Level Security Policies for Usage Events table
CREATE POLICY usage_events_tenant_isolation ON usage_events
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR is_platform_admin()
    );

-- Row Level Security Policies for Subscriptions table
CREATE POLICY subscriptions_tenant_isolation ON subscriptions
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR is_platform_admin()
    );

-- Row Level Security Policies for Invoices table
CREATE POLICY invoices_tenant_isolation ON invoices
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR is_platform_admin()
    );

-- Row Level Security Policies for Feedback table
CREATE POLICY feedback_tenant_isolation ON feedback
    FOR ALL USING (
        user_id = current_user_id()
        OR user_id IS NULL  -- Allow anonymous feedback access for the user who created it
        OR is_platform_admin()
    );

-- Row Level Security Policies for Content tables (public access for published content)
CREATE POLICY changelog_entries_public_read ON changelog_entries
    FOR SELECT USING (is_published = true OR is_platform_admin());

CREATE POLICY blog_posts_public_read ON blog_posts
    FOR SELECT USING (is_published = true OR is_platform_admin());

-- Row Level Security Policies for Monitoring tables
CREATE POLICY performance_metrics_tenant_isolation ON performance_metrics
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR tenant_id IS NULL  -- Allow system-wide metrics
        OR is_platform_admin()
    );

CREATE POLICY alerts_tenant_isolation ON alerts
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR tenant_id IS NULL  -- Allow system-wide alerts
        OR is_platform_admin()
    );

-- Row Level Security Policies for Security tables
CREATE POLICY security_scans_tenant_isolation ON security_scans
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR tenant_id IS NULL  -- Allow system-wide scans
        OR is_platform_admin()
    );

CREATE POLICY vulnerabilities_tenant_isolation ON vulnerabilities
    FOR ALL USING (
        scan_id IN (
            SELECT id FROM security_scans
            WHERE tenant_id = current_tenant_id() OR tenant_id IS NULL
        )
        OR is_platform_admin()
    );

-- Row Level Security Policies for Dashboard configs
CREATE POLICY dashboard_configs_tenant_isolation ON dashboard_configs
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR is_platform_admin()
    );

-- Row Level Security Policies for Functions
CREATE POLICY function_configs_tenant_isolation ON function_configs
    FOR ALL USING (
        tenant_id = current_tenant_id()
        OR is_platform_admin()
    );

CREATE POLICY function_deployments_tenant_isolation ON function_deployments
    FOR ALL USING (
        function_id IN (
            SELECT id FROM function_configs WHERE tenant_id = current_tenant_id()
        )
        OR is_platform_admin()
    );

CREATE POLICY function_logs_tenant_isolation ON function_logs
    FOR ALL USING (
        function_id IN (
            SELECT id FROM function_configs WHERE tenant_id = current_tenant_id()
        )
        OR function_id IS NULL  -- Allow system logs
        OR is_platform_admin()
    );

-- Create audit trigger function for automatic audit logging
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS trigger AS $$
DECLARE
    old_row jsonb;
    new_row jsonb;
    action_type text;
    resource_id uuid;
    tenant_id uuid;
BEGIN
    -- Determine action type
    IF TG_OP = 'INSERT' THEN
        action_type := 'create';
        old_row := NULL;
        new_row := row_to_json(NEW)::jsonb;
    ELSIF TG_OP = 'UPDATE' THEN
        action_type := 'update';
        old_row := row_to_json(OLD)::jsonb;
        new_row := row_to_json(NEW)::jsonb;
    ELSIF TG_OP = 'DELETE' THEN
        action_type := 'delete';
        old_row := row_to_json(OLD)::jsonb;
        new_row := NULL;
    END IF;

    -- Try to extract tenant_id from the record
    BEGIN
        -- Handle different table structures
        CASE TG_TABLE_NAME
            WHEN 'users', 'apps', 'subscriptions', 'invoices', 'usage_events', 'dashboard_configs', 'function_configs' THEN
                tenant_id := CASE
                    WHEN TG_OP = 'DELETE' THEN (old_row->>'tenant_id')::uuid
                    ELSE (new_row->>'tenant_id')::uuid
                END;
            WHEN 'backends', 'deployments' THEN
                -- For backends/deployments, we need to look up the tenant via app_id
                DECLARE
                    app_id uuid;
                BEGIN
                    app_id := CASE
                        WHEN TG_OP = 'DELETE' THEN (old_row->>'app_id')::uuid
                        ELSE (new_row->>'app_id')::uuid
                    END;
                    SELECT a.tenant_id INTO tenant_id FROM apps a WHERE a.id = app_id;
                END;
            WHEN 'function_deployments' THEN
                -- For function deployments, look up via function_id
                DECLARE
                    function_id uuid;
                BEGIN
                    function_id := CASE
                        WHEN TG_OP = 'DELETE' THEN (old_row->>'function_id')::uuid
                        ELSE (new_row->>'function_id')::uuid
                    END;
                    SELECT f.tenant_id INTO tenant_id FROM function_configs f WHERE f.id = function_id;
                END;
            ELSE
                tenant_id := NULL;
        END CASE;
    EXCEPTION WHEN OTHERS THEN
        tenant_id := NULL;
    END;

    -- Extract resource ID
    BEGIN
        resource_id := CASE
            WHEN TG_OP = 'DELETE' THEN (old_row->>'id')::uuid
            ELSE (new_row->>'id')::uuid
        END;
    EXCEPTION WHEN OTHERS THEN
        resource_id := NULL;
    END;

    -- Insert audit record
    INSERT INTO audit_events (
        actor_user_id,
        tenant_id,
        action,
        resource_type,
        resource_id,
        before_state,
        after_state,
        ip_address,
        user_agent,
        timestamp,
        success
    ) VALUES (
        current_user_id(),
        tenant_id,
        action_type || '_' || TG_TABLE_NAME,
        TG_TABLE_NAME,
        resource_id,
        old_row,
        new_row,
        NULLIF(current_setting('app.client_ip', true), ''),
        NULLIF(current_setting('app.user_agent', true), ''),
        now(),
        true
    );

    -- For INSERT and UPDATE, return NEW; for DELETE, return OLD
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create audit triggers for sensitive tables
CREATE TRIGGER audit_users_trigger
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_apps_trigger
    AFTER INSERT OR UPDATE OR DELETE ON apps
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_backends_trigger
    AFTER INSERT OR UPDATE OR DELETE ON backends
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_deployments_trigger
    AFTER INSERT OR UPDATE OR DELETE ON deployments
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_subscriptions_trigger
    AFTER INSERT OR UPDATE OR DELETE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_function_configs_trigger
    AFTER INSERT OR UPDATE OR DELETE ON function_configs
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_function_deployments_trigger
    AFTER INSERT OR UPDATE OR DELETE ON function_deployments
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Create a function to validate tenant access
CREATE OR REPLACE FUNCTION validate_tenant_access(target_tenant_id uuid)
RETURNS boolean AS $$
BEGIN
    -- Allow access if user is platform admin or belongs to the target tenant
    RETURN is_platform_admin() OR current_tenant_id() = target_tenant_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create a function to check rate limits for tenant operations
CREATE OR REPLACE FUNCTION check_tenant_rate_limit(
    operation_type text,
    max_operations integer DEFAULT 100,
    time_window interval DEFAULT '1 hour'
)
RETURNS boolean AS $$
DECLARE
    operation_count integer;
BEGIN
    -- Count recent operations by this tenant
    SELECT COUNT(*) INTO operation_count
    FROM audit_events
    WHERE tenant_id = current_tenant_id()
    AND action = operation_type
    AND timestamp > now() - time_window;

    -- Allow operation if under the limit
    RETURN operation_count < max_operations;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create a function to automatically clean up expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS integer AS $$
DECLARE
    deleted_count integer;
BEGIN
    DELETE FROM sessions WHERE expires_at < now();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create a function to validate function code security
CREATE OR REPLACE FUNCTION validate_function_security(function_code text)
RETURNS boolean AS $$
BEGIN
    -- Basic security checks - you can extend this with more sophisticated checks
    -- Reject code containing dangerous patterns
    IF function_code ~* '(eval|exec|require\s*\(\s*.*child_process|fs\.)' THEN
        RETURN false;
    END IF;

    -- Check for reasonable code length
    IF length(function_code) > 100000 THEN  -- 100KB limit
        RETURN false;
    END IF;

    RETURN true;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Add check constraint to function_configs for security validation
ALTER TABLE function_configs
ADD CONSTRAINT function_code_security_check
CHECK (validate_function_security(code));

-- Create indexes to support RLS policies efficiently
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_apps_tenant_id ON apps(tenant_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_backends_app_id ON backends(app_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_deployments_app_id ON deployments(app_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_events_tenant_timestamp ON audit_events(tenant_id, timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_usage_events_tenant_timestamp ON usage_events(tenant_id, timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_subscriptions_tenant_id ON subscriptions(tenant_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_tenant_id ON invoices(tenant_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_performance_metrics_tenant_timestamp ON performance_metrics(tenant_id, timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_security_scans_tenant_id ON security_scans(tenant_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dashboard_configs_tenant_id ON dashboard_configs(tenant_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_function_configs_tenant_id ON function_configs(tenant_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_function_deployments_function_id ON function_deployments(function_id);

-- Create a view for security monitoring
CREATE OR REPLACE VIEW security_monitoring AS
SELECT
    ae.timestamp,
    ae.action,
    ae.resource_type,
    ae.resource_id,
    u.email as actor_email,
    t.name as tenant_name,
    ae.ip_address,
    ae.user_agent,
    ae.success
FROM audit_events ae
LEFT JOIN users u ON u.id = ae.actor_user_id
LEFT JOIN tenants t ON t.id = ae.tenant_id
WHERE ae.timestamp > now() - interval '24 hours'
ORDER BY ae.timestamp DESC;

-- Create a function to get tenant security summary
CREATE OR REPLACE FUNCTION get_tenant_security_summary(target_tenant_id uuid)
RETURNS jsonb AS $$
DECLARE
    result jsonb;
BEGIN
    IF NOT validate_tenant_access(target_tenant_id) THEN
        RAISE EXCEPTION 'Access denied to tenant security data';
    END IF;

    SELECT jsonb_build_object(
        'tenant_id', target_tenant_id,
        'failed_logins_24h', (
            SELECT count(*) FROM audit_events
            WHERE tenant_id = target_tenant_id
            AND action LIKE '%login%'
            AND success = false
            AND timestamp > now() - interval '24 hours'
        ),
        'suspicious_activities_24h', (
            SELECT count(*) FROM audit_events
            WHERE tenant_id = target_tenant_id
            AND action IN ('delete_user', 'delete_app', 'delete_deployment')
            AND timestamp > now() - interval '24 hours'
        ),
        'active_sessions', (
            SELECT count(*) FROM sessions s
            JOIN users u ON u.id = s.user_id
            WHERE u.tenant_id = target_tenant_id
            AND s.expires_at > now()
        ),
        'recent_security_scans', (
            SELECT count(*) FROM security_scans
            WHERE tenant_id = target_tenant_id
            AND started_at > now() - interval '7 days'
        )
    ) INTO result;

    RETURN result;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;