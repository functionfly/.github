-- +migration File: 20260625000000_state_values_key_id.down.sql

ALTER TABLE state_values DROP COLUMN IF EXISTS key_id;
