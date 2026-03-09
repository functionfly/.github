-- Drop email_events table and related objects
DROP TRIGGER IF EXISTS email_events_updated_at ON email_events;
DROP FUNCTION IF EXISTS update_email_events_updated_at();
DROP TABLE IF EXISTS email_events;
