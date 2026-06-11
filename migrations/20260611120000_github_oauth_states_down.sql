-- Remove GitHub OAuth fields from oauth_states
ALTER TABLE oauth_states DROP CONSTRAINT IF EXISTS chk_oauth_states_user_id;
ALTER TABLE oauth_states DROP CONSTRAINT IF EXISTS chk_oauth_states_tenant_id;
ALTER TABLE oauth_states DROP COLUMN IF EXISTS provider;
ALTER TABLE oauth_states DROP COLUMN IF EXISTS user_id;
ALTER TABLE oauth_states DROP COLUMN IF EXISTS tenant_id;