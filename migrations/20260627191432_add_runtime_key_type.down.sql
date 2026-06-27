-- Revert key_type CHECK constraint to pre-runtime types

ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_key_type_check;

ALTER TABLE api_keys ADD CONSTRAINT api_keys_key_type_check
    CHECK (key_type IN ('platform', 'function', 'agent', 'environment', 'oauth', 'trust'));
