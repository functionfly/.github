-- Migration: 20260323000000_api_keys_unique_active_only
-- Description: Rollback - restore original unique constraint on (tenant_id, name) for all rows.
-- Created: 2026-03-23
-- Author: FunctionFly

DROP INDEX IF EXISTS idx_api_keys_tenant_name_active;

-- Restore original constraint (may fail if duplicate names exist across active+inactive)
ALTER TABLE api_keys ADD CONSTRAINT tenant_api_keys_unique UNIQUE (tenant_id, name);
