-- Rollback audit trigger
-- This removes all audit triggers and the audit log table

-- Remove all audit triggers from tables
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT trigger_name, event_object_table
        FROM information_schema.triggers
        WHERE trigger_name LIKE 'audit_%'
        AND event_object_schema = 'public'
    LOOP
        EXECUTE format('DROP TRIGGER IF EXISTS %I ON %I', rec.trigger_name, rec.event_object_table);
        RAISE NOTICE 'Dropped trigger % from table %', rec.trigger_name, rec.event_object_table;
    END LOOP;
END $$;

-- Drop trigger functions
DROP FUNCTION IF EXISTS audit_trigger_row_changes();
DROP FUNCTION IF EXISTS audit_trigger_capture_session();
DROP FUNCTION IF EXISTS install_audit_trigger(TEXT);
DROP FUNCTION IF EXISTS install_audit_trigger_for_schema(TEXT);
DROP FUNCTION IF EXISTS uninstall_audit_trigger(TEXT);
DROP FUNCTION IF EXISTS get_audit_log(TEXT, TEXT, INTEGER);
DROP FUNCTION IF EXISTS cleanup_audit_trigger_log(INTEGER);

-- Drop audit log table
DROP TABLE IF EXISTS audit_trigger_log;

-- Note: This down migration is destructive and will remove all audit data.
-- In production, consider archiving the audit log before applying this migration.
