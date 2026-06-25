-- +go migration
-- Remove consciousness analyzer columns

BEGIN;

ALTER TABLE usage_events DROP COLUMN IF EXISTS function_id;
ALTER TABLE function_dna_profiles DROP COLUMN IF EXISTS cold_start_rate;

COMMIT;
