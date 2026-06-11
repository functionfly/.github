-- Rollback migration state transition
-- This restores the old custom migration table format

-- Drop the golang-migrate table
DROP TABLE IF EXISTS schema_migrations;

-- Restore old table if it exists
DO $$
BEGIN
    -- Check if backup table exists
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = 'schema_migrations_old'
        AND table_schema = 'public'
    ) THEN
        -- Restore old table
        ALTER TABLE schema_migrations_old RENAME TO schema_migrations;
        RAISE NOTICE 'Restored old schema_migrations table';
    ELSE
        RAISE NOTICE 'No backup table found, creating empty old-format table';
        -- Create old format table as fallback
        CREATE TABLE schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
    END IF;
END $$;