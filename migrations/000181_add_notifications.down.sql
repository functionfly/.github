-- Drop triggers
DROP TRIGGER IF EXISTS trigger_notifications_updated_at ON notifications;
DROP TRIGGER IF EXISTS trigger_notification_preferences_updated_at ON notification_preferences;
DROP TRIGGER IF EXISTS trigger_notification_templates_updated_at ON notification_templates;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_notifications_updated_at();

-- Drop tables (in reverse order of creation due to foreign key constraints)
DROP TABLE IF EXISTS notification_analytics;
DROP TABLE IF EXISTS notification_templates;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS notifications;
