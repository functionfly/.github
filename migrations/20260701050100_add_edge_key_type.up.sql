-- Migration: 20260701050100_add_edge_key_type
-- Description: Add 'edge' to api_keys key_type check constraint for FunctionFly Edge provider keys
-- Created: 2026-07-01

ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_key_type_check;
ALTER TABLE api_keys ADD CONSTRAINT api_keys_key_type_check
    CHECK (key_type IN ('platform', 'function', 'agent', 'environment', 'oauth', 'trust', 'edge'));
