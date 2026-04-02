-- Restores previous notify_database_change (matches 000122); breaks backends INSERT again.
CREATE OR REPLACE FUNCTION notify_database_change()
RETURNS TRIGGER AS $$
DECLARE
    payload JSONB;
    channel_name TEXT;
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

    IF TG_TABLE_NAME IN ('users', 'apps', 'backends', 'alerts', 'performance_metrics', 'monitoring_events', 'user_notifications') THEN
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
