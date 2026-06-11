-- Migration: Rollback DRE 2.0 tables
-- Drops all DRE 2.0 related tables and columns

-- ============================================
-- DROP TABLES
-- ============================================

-- Drop execution MEG records
DROP TABLE IF EXISTS execution_meg_records;

-- Drop execution certificates
DROP TABLE IF EXISTS execution_certificates;

-- Drop drift reports
DROP TABLE IF EXISTS drift_reports;

-- Drop execution passports
DROP TABLE IF EXISTS function_execution_passports;

-- Drop resource hash history
DROP TABLE IF EXISTS resource_hash_history;

-- ============================================
-- DROP COLUMNS FROM REGISTRY_FUNCTION_RATINGS
-- ============================================

ALTER TABLE registry_function_ratings
    DROP COLUMN IF EXISTS determinism_score,
    DROP COLUMN IF EXISTS replay_integrity_score,
    DROP COLUMN IF EXISTS performance_stability_score,
    DROP COLUMN IF EXISTS drift_score,
    DROP COLUMN IF EXISTS trust_score_v2,
    DROP COLUMN IF EXISTS trust_v2_updated_at;
