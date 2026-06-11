-- Rollback database change notifications
-- Remove triggers and functions for database change notifications

-- =============================================
-- DROP TRIGGERS
-- =============================================

DROP TRIGGER IF EXISTS notify_users_change ON users;
DROP TRIGGER IF EXISTS notify_apps_change ON apps;
DROP TRIGGER IF EXISTS notify_backends_change ON backends;
DROP TRIGGER IF EXISTS notify_alerts_change ON alerts;
DROP TRIGGER IF EXISTS notify_performance_metrics_change ON performance_metrics;
DROP TRIGGER IF EXISTS notify_monitoring_events_change ON monitoring_events;

-- Drop user_notifications trigger if table exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_notifications') THEN
        EXECUTE 'DROP TRIGGER IF EXISTS notify_user_notifications_change ON user_notifications';
    END IF;
END $$;

-- =============================================
-- DROP FUNCTION
-- =============================================

DROP FUNCTION IF EXISTS notify_database_change();

-- =============================================
-- DROP INDEXES (optional - keep for performance if other triggers exist)
-- =============================================

-- Note: We're keeping the indexes as they may be used by other parts of the system