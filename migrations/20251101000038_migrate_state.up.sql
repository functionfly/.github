-- Migration to handle transition from custom migration system to golang-migrate
-- This migration ensures the schema_migrations table is in the correct format

-- Check if we need to migrate from old format
DO $$
DECLARE
    old_table_exists BOOLEAN;
    max_version BIGINT := 0;
BEGIN
    -- Check if the old schema_migrations table exists with old structure
    SELECT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'schema_migrations'
        AND column_name = 'applied_at'
        AND table_schema = 'public'
    ) INTO old_table_exists;

    IF old_table_exists THEN
        -- Migrate from old format (version VARCHAR, applied_at TIMESTAMP)
        -- to new format (version BIGINT, dirty BOOLEAN)

        RAISE NOTICE 'Migrating from old schema_migrations format...';

        -- Get the maximum version
        SELECT COALESCE(MAX(CAST(version AS BIGINT)), 0)
        INTO max_version
        FROM schema_migrations;

        -- Create new table with correct structure
        CREATE TABLE schema_migrations_new (
            version BIGINT NOT NULL PRIMARY KEY,
            dirty BOOLEAN NOT NULL DEFAULT FALSE
        );

        -- Insert migrated data
        INSERT INTO schema_migrations_new (version, dirty)
        VALUES (max_version, FALSE);

        -- Backup old table and replace
        ALTER TABLE schema_migrations RENAME TO schema_migrations_old;
        ALTER TABLE schema_migrations_new RENAME TO schema_migrations;

        RAISE NOTICE 'Migration state migrated successfully from version %', max_version;
    ELSE
        -- Ensure the table exists with correct structure
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version BIGINT NOT NULL PRIMARY KEY,
            dirty BOOLEAN NOT NULL DEFAULT FALSE
        );

        RAISE NOTICE 'Created new schema_migrations table for golang-migrate';
    END IF;
END $$;