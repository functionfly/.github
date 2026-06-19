-- Revert: Remove automated partition creation cron job
BEGIN;

DELETE FROM cron.job WHERE jobname = 'create-next-month-partitions';

COMMIT;
