-- Rollback: drop TimescaleDB rebuild objects
-- Note: This is a best-effort rollback - some objects may need manual cleanup

BEGIN;

-- Drop pg_cron jobs
DO $$
DECLARE
    job record;
BEGIN
    FOR job IN SELECT jobname FROM cron.job WHERE jobname IN ('refresh-signal-stats', 'cleanup-old-signals', 'cleanup-old-analytics', 'create-next-partitions')
    LOOP
        PERFORM cron.unschedule(job.jobname);
    END LOOP;
END $$;

-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS signal_stats_hourly CASCADE;

-- Drop partitioned tables (will cascade to partitions)
DROP TABLE IF EXISTS signals CASCADE;
DROP TABLE IF EXISTS analytics_events CASCADE;

-- Extensions are not dropped as other objects may depend on them

COMMIT;