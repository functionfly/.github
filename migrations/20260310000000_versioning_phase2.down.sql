-- Rollback Versioning System Phase 2
-- Migration: 20260310000000_versioning_phase2

-- Drop rollback records table
DROP TABLE IF EXISTS rollback_records;

-- Drop version aliases table
DROP TABLE IF EXISTS version_aliases;

-- Remove columns from registry_function_versions
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS sunset_at;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS published_at;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS archived_at;

-- Remove is_default from api_versions
ALTER TABLE api_versions DROP COLUMN IF EXISTS is_default;
