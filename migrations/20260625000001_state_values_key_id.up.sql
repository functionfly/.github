-- +migration File: 20260625000000_state_values_key_id.up.sql
-- Add key_id column to state_values for multi-key encryption support
-- This enables proper key rotation tracking and multi-key decryption

ALTER TABLE state_values ADD COLUMN IF NOT EXISTS key_id VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_state_values_key_id ON state_values(key_id) WHERE key_id IS NOT NULL;

-- +migration File: 20260625000000_state_values_key_id.down.sql

ALTER TABLE state_values DROP COLUMN IF EXISTS key_id;
