ALTER TABLE oauth_states DROP COLUMN IF EXISTS invite_code;
DROP TABLE IF EXISTS signup_invite_codes;
