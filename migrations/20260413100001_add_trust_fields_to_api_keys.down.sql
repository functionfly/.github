-- Rollback: Remove Trust API fields from the unified api_keys table

-- Drop indexes
DROP INDEX IF EXISTS idx_api_keys_revoked;
DROP INDEX IF EXISTS idx_api_keys_partner;
DROP INDEX IF EXISTS idx_api_keys_key_id;

-- Remove new columns
ALTER TABLE api_keys DROP COLUMN IF EXISTS partner_id;
ALTER TABLE api_keys DROP COLUMN IF EXISTS scopes;
ALTER TABLE api_keys DROP COLUMN IF EXISTS is_revoked;
ALTER TABLE api_keys DROP COLUMN IF EXISTS revoked_at;
ALTER TABLE api_keys DROP COLUMN IF EXISTS revoked_reason;
ALTER TABLE api_keys DROP COLUMN IF EXISTS use_count;
ALTER TABLE api_keys DROP COLUMN IF EXISTS allowed_ips;
ALTER TABLE api_keys DROP COLUMN IF EXISTS created_by;
ALTER TABLE api_keys DROP COLUMN IF EXISTS key_id;

-- Restore original key_type check constraint
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_key_type_check;
ALTER TABLE api_keys ADD CONSTRAINT api_keys_key_type_check
    CHECK (key_type IN ('platform', 'function', 'agent', 'environment', 'oauth'));
