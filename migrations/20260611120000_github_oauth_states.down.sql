-- Revert: Remove GitHub OAuth columns from oauth_states
DROP INDEX IF EXISTS idx_oauth_states_user_id;

ALTER TABLE oauth_states DROP CONSTRAINT IF EXISTS chk_oauth_states_tenant_id;
ALTER TABLE oauth_states DROP CONSTRAINT IF EXISTS chk_oauth_states_user_id;
ALTER TABLE oauth_states DROP COLUMN IF EXISTS provider;
ALTER TABLE oauth_states DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE oauth_states DROP COLUMN IF EXISTS user_id;
