-- Backends rows have app_id, not tenant_id. notify_database_change() incorrectly listed
-- 'backends' alongside tables that expose tenant_id, causing:
--   ERROR: record "new" has no field "tenant_id" (SQLSTATE 42703)
-- on INSERT into backends.
--
-- Apply on Neon (or any Postgres) with psql, e.g.:
--   psql "$DATABASE_URL" -f migrations/20260328120000_fix_notify_database_change_backends.up.sql

CREATE OR REPLACE FUNCTION notify_database_change()
RETURNS TRIGGER AS $$
DECLARE
    payload JSONB;
    channel_name TEXT;
    tid uuid;
BEGIN
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

    channel_name := 'db_changes_' || TG_TABLE_NAME;

    IF TG_TABLE_NAME = 'backends' THEN
        tid := NULL;
        IF TG_OP = 'DELETE' THEN
            SELECT a.tenant_id INTO tid FROM apps a WHERE a.id = OLD.app_id;
        ELSE
            SELECT a.tenant_id INTO tid FROM apps a WHERE a.id = NEW.app_id;
        END IF;
        IF tid IS NOT NULL THEN
            channel_name := channel_name || '_' || tid::text;
        END IF;
    ELSIF TG_TABLE_NAME IN (
        'users', 'apps', 'alerts', 'performance_metrics',
        'monitoring_events', 'user_notifications'
    ) THEN
        IF TG_OP != 'DELETE' AND NEW.tenant_id IS NOT NULL THEN
            channel_name := channel_name || '_' || NEW.tenant_id::text;
        ELSIF TG_OP = 'DELETE' AND OLD.tenant_id IS NOT NULL THEN
            channel_name := channel_name || '_' || OLD.tenant_id::text;
        END IF;
    END IF;

    PERFORM pg_notify(channel_name, payload::text);

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;
