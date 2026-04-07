-- Migration: 20260401000000_verification_pipeline
-- Description: Verification Pipeline - Rollback

BEGIN;

-- Drop tables in reverse order of creation

DROP TABLE IF EXISTS verification_schedule CASCADE;
DROP TABLE IF EXISTS verification_audit_log CASCADE;
DROP TABLE IF EXISTS manual_review_queue CASCADE;
DROP TABLE IF EXISTS verification_level_config CASCADE;
DROP TABLE IF EXISTS verification_results CASCADE;
DROP TABLE IF EXISTS verification_jobs CASCADE;

-- Remove verification_level column from registry_functions
ALTER TABLE registry_functions
    DROP COLUMN IF EXISTS verification_level,
    DROP COLUMN IF EXISTS verified_at,
    DROP COLUMN IF EXISTS last_verified_at;

-- Remove verification_level from trust_history if it was added
ALTER TABLE trust_history
    DROP COLUMN IF EXISTS verification_level;

COMMIT;
