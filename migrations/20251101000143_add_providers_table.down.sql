-- Migration rollback: Remove providers table
-- Created: 2026-02-27

DROP INDEX IF EXISTS idx_providers_status;
DROP INDEX IF EXISTS idx_providers_team_id;
DROP INDEX IF EXISTS idx_providers_provider;
DROP INDEX IF EXISTS idx_providers_user_id;

DROP TABLE IF EXISTS providers;
