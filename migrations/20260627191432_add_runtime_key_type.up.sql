-- Expand the key_type CHECK constraint to include 'runtime' and 'micropython'
-- (micropython was added to the Go code but never had a CHECK migration)

ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_key_type_check;

ALTER TABLE api_keys ADD CONSTRAINT api_keys_key_type_check
    CHECK (key_type IN ('platform', 'function', 'agent', 'environment', 'oauth', 'trust', 'micropython', 'runtime'));
