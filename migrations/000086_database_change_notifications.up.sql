-- Database change notifications for real-time WebSocket updates
-- Creates triggers and functions to capture INSERT/UPDATE/DELETE operations
-- and broadcast them via PostgreSQL NOTIFY

-- =============================================
-- DATABASE CHANGE NOTIFICATION FUNCTIONS
-- =============================================

-- Function to handle database change notifications
CREATE OR REPLACE FUNCTION notify_database_change()
RETURNS TRIGGER AS $$
DECLARE
    payload JSONB;
    channel_name TEXT;
BEGIN
    -- Build the payload with change details
    payload := jsonb_build_object(
        'schema', TG_TABLE_SCHEMA,
        'table', TG_TABLE_NAME,
        'eventType', TG_OP,
        'commit_timestamp', extract(epoch from now())::text,
        'new', CASE WHEN TG_OP != 'DELETE' THEN row_to_json(NEW)::jsonb ELSE NULL END,
        'old', CASE WHEN TG_OP != 'INSERT' THEN row_to_json(OLD)::jsonb ELSE NULL END,
        'ids', CASE
            WHEN TG_OP = 'DELETE' THEN jsonb_build_array(OLD.id)
            WHEN TG_OP = 'INSERT' THEN jsonb_build_array(NEW.id)
            ELSE jsonb_build_array(COALESCE(NEW.id, OLD.id))
        END,
        'errors', NULL
    );

    -- Determine channel name based on table and tenant (if applicable)
    channel_name := 'db_changes_' || TG_TABLE_NAME;

    -- Add tenant-specific suffix if tenant_id exists
    IF TG_TABLE_NAME IN ('users', 'apps', 'backends', 'alerts', 'performance_metrics', 'monitoring_events', 'user_notifications') THEN
        IF TG_OP != 'DELETE' AND NEW.tenant_id IS NOT NULL THEN
            channel_name := channel_name || '_' || NEW.tenant_id::text;
        ELSIF TG_OP = 'DELETE' AND OLD.tenant_id IS NOT NULL THEN
            channel_name := channel_name || '_' || OLD.tenant_id::text;
        END IF;
    END IF;

    -- Send notification
    PERFORM pg_notify(channel_name, payload::text);

    -- Return appropriate row based on operation
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- TRIGGERS FOR TABLES THAT NEED REAL-TIME UPDATES
-- =============================================

-- Users table (for profile updates and user status changes)
DROP TRIGGER IF EXISTS notify_users_change ON users;
CREATE TRIGGER notify_users_change
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION notify_database_change();

-- Apps table (for app creation/updates/deletion)
DROP TRIGGER IF EXISTS notify_apps_change ON apps;
CREATE TRIGGER notify_apps_change
    AFTER INSERT OR UPDATE OR DELETE ON apps
    FOR EACH ROW EXECUTE FUNCTION notify_database_change();

-- Backends table (for backend changes)
DROP TRIGGER IF EXISTS notify_backends_change ON backends;
CREATE TRIGGER notify_backends_change
    AFTER INSERT OR UPDATE OR DELETE ON backends
    FOR EACH ROW EXECUTE FUNCTION notify_database_change();

-- Alerts table (for real-time alert notifications)
DROP TRIGGER IF EXISTS notify_alerts_change ON alerts;
CREATE TRIGGER notify_alerts_change
    AFTER INSERT OR UPDATE OR DELETE ON alerts
    FOR EACH ROW EXECUTE FUNCTION notify_database_change();

-- Performance metrics table (for real-time metrics updates)
DROP TRIGGER IF EXISTS notify_performance_metrics_change ON performance_metrics;
CREATE TRIGGER notify_performance_metrics_change
    AFTER INSERT OR UPDATE OR DELETE ON performance_metrics
    FOR EACH ROW EXECUTE FUNCTION notify_database_change();

-- Monitoring events table (for real-time monitoring events)
DROP TRIGGER IF EXISTS notify_monitoring_events_change ON monitoring_events;
CREATE TRIGGER notify_monitoring_events_change
    AFTER INSERT OR UPDATE OR DELETE ON monitoring_events
    FOR EACH ROW EXECUTE FUNCTION notify_database_change();

-- User notifications table (for real-time notification updates)
-- First check if the table exists (might not exist in all deployments)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_notifications') THEN
        EXECUTE 'DROP TRIGGER IF EXISTS notify_user_notifications_change ON user_notifications';
        EXECUTE 'CREATE TRIGGER notify_user_notifications_change
            AFTER INSERT OR UPDATE OR DELETE ON user_notifications
            FOR EACH ROW EXECUTE FUNCTION notify_database_change()';
    END IF;
END $$;

-- =============================================
-- INDEXES FOR PERFORMANCE
-- =============================================

-- Add indexes on commonly filtered columns for better performance
CREATE INDEX IF NOT EXISTS idx_users_tenant_updated ON users(tenant_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_apps_tenant_updated ON apps(tenant_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_backends_app_updated ON backends(app_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_tenant_status ON alerts(tenant_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_tenant_type ON performance_metrics(tenant_id, metric_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_monitoring_events_tenant_type ON monitoring_events(tenant_id, event_type, timestamp DESC);