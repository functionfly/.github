-- Down migration for profile_number column

DROP INDEX IF EXISTS idx_users_profile_number;

ALTER TABLE users DROP COLUMN IF EXISTS profile_number;

DROP SEQUENCE IF EXISTS users_profile_number_seq;
