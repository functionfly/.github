BEGIN;

DROP TABLE IF EXISTS team_quotas;
DROP TABLE IF EXISTS team_audit_log;
DROP TABLE IF EXISTS team_api_keys;
ALTER TABLE teams DROP COLUMN IF EXISTS default_invite_role;

COMMIT;
