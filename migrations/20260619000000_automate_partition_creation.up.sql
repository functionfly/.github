-- Migration: 20260619000000_automate_partition_creation
-- Description: Add pg_cron job to automatically create future partitions
-- for registry_function_executions, performance_metrics, system_health_checks,
-- database_metrics, and function_logs tables

BEGIN;

-- ============================================================
-- pg_cron job: create-next-month-partitions
-- Runs on the 25th of each month to create partitions 2 months ahead
-- This ensures partitions exist before data arrives
-- ============================================================

-- Drop existing job if it exists (idempotent)
DELETE FROM cron.job WHERE jobname = 'create-next-month-partitions';

-- Schedule: 25th of every month at midnight
SELECT cron.schedule(
    'create-next-month-partitions',
    '0 0 25 * *',
    $$SELECT create_next_month_partitions()$$
);

-- Verify the job was created
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM cron.job WHERE jobname = 'create-next-month-partitions') THEN
        RAISE NOTICE 'Successfully scheduled create-next-month-partitions cron job';
    ELSE
        RAISE WARNING 'Failed to schedule create-next-month-partitions cron job - ensure pg_cron is configured';
    END IF;
END $$;

-- Immediately create partitions for the next 3 months (no wait for cron)
DO $$
BEGIN
    PERFORM create_next_month_partitions();
    RAISE NOTICE 'Immediately created near-term partitions';
END $$;

COMMIT;
