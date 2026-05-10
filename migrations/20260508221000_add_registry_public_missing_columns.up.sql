-- Migration: Add missing columns to registry_executions_public
-- The Go model expects verified_at, verification_status, verification_error,
-- replayed_output_json, and replayed_duration_ms but these were never added to the table.

ALTER TABLE registry_executions_public
    ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS verification_status VARCHAR(20),
    ADD COLUMN IF NOT EXISTS verification_error TEXT,
    ADD COLUMN IF NOT EXISTS replayed_output_json JSONB,
    ADD COLUMN IF NOT EXISTS replayed_duration_ms INTEGER;