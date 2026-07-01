BEGIN;

ALTER TABLE api_keys DROP COLUMN IF EXISTS team_id;
-- team_api_keys is not restored here since it was dropped in the up migration

COMMIT;
