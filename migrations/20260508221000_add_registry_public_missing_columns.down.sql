-- Rollback: Remove added columns
ALTER TABLE registry_executions_public
    DROP COLUMN IF EXISTS verified_at,
    DROP COLUMN IF EXISTS verification_status,
    DROP COLUMN IF EXISTS verification_error,
    DROP COLUMN IF EXISTS replayed_output_json,
    DROP COLUMN IF EXISTS replayed_duration_ms;