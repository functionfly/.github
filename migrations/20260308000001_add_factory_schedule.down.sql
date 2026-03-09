-- Remove scheduling fields from factory_config table
ALTER TABLE factory_config
DROP COLUMN IF EXISTS schedule_enabled,
DROP COLUMN IF EXISTS schedule_cron,
DROP COLUMN IF EXISTS schedule_timezone;
