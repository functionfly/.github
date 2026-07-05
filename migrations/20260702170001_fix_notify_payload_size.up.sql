-- Fix: Limit the payload size sent via pg_notify to avoid "payload string too long" errors
-- PostgreSQL's NOTIFY has an 8000-byte limit on the payload parameter.
-- This fix truncates the 'new' and 'old' row data to prevent exceeding this limit
-- while still preserving the IDs and essential change information.

CREATE OR REPLACE FUNCTION notify_database_change()
RETURNS TRIGGER AS $$
DECLARE
    payload JSONB;
    channel_name TEXT;
    payload_text TEXT;
    max_payload_size INTEGER := 7000; -- Leave some buffer under 8000 limit
BEGIN
    -- Build the payload with change details, but limit row data size
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

    -- Convert to text and truncate if necessary
    payload_text := payload::text;
    IF length(payload_text) > max_payload_size THEN
        -- Truncate and add indicator that data was truncated
        payload := jsonb_build_object(
            'schema', TG_TABLE_SCHEMA,
            'table', TG_TABLE_NAME,
            'eventType', TG_OP,
            'commit_timestamp', extract(epoch from now())::text,
            'new', CASE WHEN TG_OP != 'DELETE' THEN jsonb_build_object('id', NEW.id, '_truncated', true) ELSE NULL END,
            'old', CASE WHEN TG_OP != 'INSERT' THEN jsonb_build_object('id', OLD.id, '_truncated', true) ELSE NULL END,
            'ids', CASE
                WHEN TG_OP = 'DELETE' THEN jsonb_build_array(OLD.id)
                WHEN TG_OP = 'INSERT' THEN jsonb_build_array(NEW.id)
                ELSE jsonb_build_array(COALESCE(NEW.id, OLD.id))
            END,
            'errors', 'payload_truncated',
            '_original_size', length(payload_text)
        );
        payload_text := payload::text;
    END IF;

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

    -- Send notification (payload is now guaranteed to be under limit)
    PERFORM pg_notify(channel_name, payload_text);

    -- Return appropriate row based on operation
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;
