-- Migration: Fix geo_country column size to support regional codes up to 20 chars
-- The previous VARCHAR(2) was too small for values like "private", "unknown", or future regional codes
--
-- Previous: geo_country VARCHAR(2)
-- Now: geo_country VARCHAR(20)

ALTER TABLE registry_function_executions
    ALTER COLUMN geo_country TYPE VARCHAR(20);

-- Also update any partitioned tables if they exist
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'registry_function_executions_partitioned') THEN
        ALTER TABLE registry_function_executions_partitioned
            ALTER COLUMN geo_country TYPE VARCHAR(20);
    END IF;
EXCEPTION WHEN OTHERS THEN
    -- Ignore if table doesn't exist or other errors
    RAISE NOTICE 'Partitioned table may not exist or already migrated: %', SQLERRM;
END $$;