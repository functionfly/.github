-- Audit Trigger for Database Changes
-- This trigger captures all INSERT, UPDATE, DELETE operations on tracked tables
-- and logs them to the audit_events table with the session user and timestamp.
--
-- Usage:
--   SELECT install_audit_trigger('tenants');
--   SELECT install_audit_trigger('users');
--   etc.
--
-- To track all tables in a schema:
--   SELECT install_audit_trigger_for_schema('public');

-- Create the audit log table if it doesn't exist
CREATE TABLE IF NOT EXISTS audit_trigger_log (
    id              BIGSERIAL PRIMARY KEY,
    session_user    TEXT NOT NULL,
    action          TEXT NOT NULL,
    table_name      TEXT NOT NULL,
    row_id          TEXT,
    changes         JSONB,
    executed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    client_addr     INET,
    application_name TEXT
);

-- Index for querying by table and time
CREATE INDEX IF NOT EXISTS idx_audit_trigger_log_table_time
    ON audit_trigger_log(table_name, executed_at DESC);

-- Index for querying by action type
CREATE INDEX IF NOT EXISTS idx_audit_trigger_log_action
    ON audit_trigger_log(action, executed_at DESC);

-- Function to capture session info
CREATE OR REPLACE FUNCTION audit_trigger_capture_session()
RETURNS JSONB AS $$
DECLARE
    result JSONB;
    client_addr inet;
    app_name text;
BEGIN
    -- Get the client address if available
    BEGIN
        client_addr := inet_client_addr();
    EXCEPTION WHEN OTHERS THEN
        client_addr := NULL;
    END;

    BEGIN
        app_name := current_setting('application_name', TRUE);
    EXCEPTION WHEN OTHERS THEN
        app_name := NULL;
    END;

    result := jsonb_build_object(
        'session_user', current_user,
        'client_addr', CASE WHEN client_addr IS NOT NULL THEN client_addr::text ELSE NULL END,
        'application_name', app_name
    );

    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Drop existing audit trigger function if it exists
DROP FUNCTION IF EXISTS audit_trigger_row_changes();

-- Function to capture row changes
CREATE OR REPLACE FUNCTION audit_trigger_row_changes()
RETURNS TRIGGER AS $$
DECLARE
    audit_entry JSONB;
    session_info JSONB;
BEGIN
    -- Capture session information
    session_info := audit_trigger_capture_session();

    IF TG_OP = 'INSERT' THEN
        audit_entry := jsonb_build_object(
            'action', 'INSERT',
            'table', TG_TABLE_NAME,
            'schema', TG_TABLE_SCHEMA,
            'new_row', to_jsonb(NEW),
            'session', session_info
        );

        -- For row-level logging, insert into the audit log
        INSERT INTO audit_trigger_log (session_user, action, table_name, row_id, changes, client_addr, application_name)
        VALUES (
            session_info->>'session_user',
            'INSERT',
            TG_TABLE_NAME,
            NEW.id::text,
            jsonb_build_object('new', to_jsonb(NEW)),
            CASE WHEN session_info->>'client_addr' IS NOT NULL
                 THEN (session_info->>'client_addr')::inet
                 ELSE NULL END,
            session_info->>'application_name'
        );

        RETURN NEW;

    ELSIF TG_OP = 'UPDATE' THEN
        audit_entry := jsonb_build_object(
            'action', 'UPDATE',
            'table', TG_TABLE_NAME,
            'schema', TG_TABLE_SCHEMA,
            'old_row', to_jsonb(OLD),
            'new_row', to_jsonb(NEW),
            'session', session_info
        );

        INSERT INTO audit_trigger_log (session_user, action, table_name, row_id, changes, client_addr, application_name)
        VALUES (
            session_info->>'session_user',
            'UPDATE',
            TG_TABLE_NAME,
            NEW.id::text,
            jsonb_build_object('old', to_jsonb(OLD), 'new', to_jsonb(NEW)),
            CASE WHEN session_info->>'client_addr' IS NOT NULL
                 THEN (session_info->>'client_addr')::inet
                 ELSE NULL END,
            session_info->>'application_name'
        );

        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        audit_entry := jsonb_build_object(
            'action', 'DELETE',
            'table', TG_TABLE_NAME,
            'schema', TG_TABLE_SCHEMA,
            'old_row', to_jsonb(OLD),
            'session', session_info
        );

        INSERT INTO audit_trigger_log (session_user, action, table_name, row_id, changes, client_addr, application_name)
        VALUES (
            session_info->>'session_user',
            'DELETE',
            TG_TABLE_NAME,
            OLD.id::text,
            jsonb_build_object('old', to_jsonb(OLD)),
            CASE WHEN session_info->>'client_addr' IS NOT NULL
                 THEN (session_info->>'client_addr')::inet
                 ELSE NULL END,
            session_info->>'application_name'
        );

        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Drop existing install function if exists
DROP FUNCTION IF EXISTS install_audit_trigger(text);

-- Function to install audit trigger on a table
CREATE OR REPLACE FUNCTION install_audit_trigger(table_name TEXT)
RETURNS void AS $$
DECLARE
    trigger_name TEXT;
    id_column_exists boolean;
BEGIN
    trigger_name := 'audit_' || table_name || '_row_changes';

    -- Check if table exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = install_audit_trigger.table_name
    ) THEN
        RAISE EXCEPTION 'Table % does not exist', table_name;
    END IF;

    -- Check if table has an 'id' column (we'll use that as the row identifier)
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
        AND table_name = install_audit_trigger.table_name
        AND column_name = 'id'
    ) INTO id_column_exists;

    IF NOT id_column_exists THEN
        RAISE WARNING 'Table % does not have an id column, row identification may be incomplete', table_name;
    END IF;

    -- Drop existing trigger if it exists
    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', trigger_name, table_name);

    -- Create the trigger
    EXECUTE format(
        'CREATE TRIGGER %I AFTER INSERT OR UPDATE OR DELETE ON %I
         FOR EACH ROW EXECUTE FUNCTION audit_trigger_row_changes()',
        trigger_name, table_name
    );

    RAISE NOTICE 'Audit trigger % installed on table %', trigger_name, table_name;
END;
$$ LANGUAGE plpgsql;

-- Drop existing install function for schema
DROP FUNCTION IF EXISTS install_audit_trigger_for_schema(text);

-- Function to install audit triggers on all tables in a schema
CREATE OR REPLACE FUNCTION install_audit_trigger_for_schema(schema_name TEXT DEFAULT 'public')
RETURNS void AS $$
DECLARE
    rec RECORD;
    tables_installed integer := 0;
    tables_skipped integer := 0;
BEGIN
    FOR rec IN
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = schema_name
        AND table_type = 'BASE TABLE'
        -- Skip audit and system tables
        AND table_name NOT LIKE 'audit_%'
        AND table_name NOT LIKE 'pg_%'
        AND table_name != 'schema_migrations'
        AND table_name != 'flyway_schema_history'
        AND table_name != 'golang_migrations'
    LOOP
        BEGIN
            PERFORM install_audit_trigger(rec.table_name);
            tables_installed := tables_installed + 1;
        EXCEPTION WHEN OTHERS THEN
            RAISE WARNING 'Could not install trigger on table %: %', rec.table_name, SQLERRM;
            tables_skipped := tables_skipped + 1;
        END;
    END LOOP;

    RAISE NOTICE 'Installed audit triggers on % tables, skipped % tables', tables_installed, tables_skipped;
END;
$$ LANGUAGE plpgsql;

-- Drop existing uninstall function
DROP FUNCTION IF EXISTS uninstall_audit_trigger(text);

-- Function to remove audit trigger from a table
CREATE OR REPLACE FUNCTION uninstall_audit_trigger(table_name TEXT)
RETURNS void AS $$
DECLARE
    trigger_name TEXT;
BEGIN
    trigger_name := 'audit_' || table_name || '_row_changes';

    EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', trigger_name, table_name);

    RAISE NOTICE 'Audit trigger % removed from table %', trigger_name, table_name;
END;
$$ LANGUAGE plpgsql;

-- Function to get audit log for a specific table and row
DROP FUNCTION IF EXISTS get_audit_log(text, text, integer);

CREATE OR REPLACE FUNCTION get_audit_log(
    p_table_name TEXT,
    p_row_id TEXT DEFAULT NULL,
    p_limit INTEGER DEFAULT 100
)
RETURNS TABLE (
    id            BIGINT,
    session_user  TEXT,
    action        TEXT,
    table_name    TEXT,
    row_id        TEXT,
    changes       JSONB,
    executed_at   TIMESTAMPTZ,
    client_addr   INET,
    app_name      TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        atl.id,
        atl.session_user,
        atl.action,
        atl.table_name,
        atl.row_id,
        atl.changes,
        atl.executed_at,
        atl.client_addr,
        atl.application_name
    FROM audit_trigger_log atl
    WHERE atl.table_name = p_table_name
    AND (p_row_id IS NULL OR atl.row_id = p_row_id)
    ORDER BY atl.executed_at DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Retention policy for audit trigger log (30 days default)
-- This can be adjusted based on compliance requirements
DROP FUNCTION IF EXISTS cleanup_audit_trigger_log(integer);

CREATE OR REPLACE FUNCTION cleanup_audit_trigger_log(retention_days INTEGER DEFAULT 30)
RETURNS integer AS $$
DECLARE
    deleted_count integer;
    cutoff_date timestamptz;
BEGIN
    cutoff_date := NOW() - (retention_days || ' days')::interval;

    DELETE FROM audit_trigger_log
    WHERE executed_at < cutoff_date;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Comment explaining usage
COMMENT ON TABLE audit_trigger_log IS 'Automatic audit log for database changes captured via triggers. Provides defense-in-depth when application-level audit logging is bypassed.';
COMMENT ON FUNCTION audit_trigger_row_changes() IS 'Trigger function that captures INSERT/UPDATE/DELETE operations and logs them to audit_trigger_log.';
COMMENT ON FUNCTION install_audit_trigger(TEXT) IS 'Installs audit trigger on a specific table. Usage: SELECT install_audit_trigger(''users'');';
COMMENT ON FUNCTION install_audit_trigger_for_schema(TEXT) IS 'Installs audit triggers on all tables in a schema. Usage: SELECT install_audit_trigger_for_schema(''public'');';
