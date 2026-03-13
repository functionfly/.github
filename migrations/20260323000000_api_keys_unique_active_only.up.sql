-- Migration: 20260323000000_api_keys_unique_active_only
-- Description: Allow reusing API key name after soft-delete by enforcing uniqueness only for active keys.
-- Created: 2026-03-23
-- Author: FunctionFly

-- Drop the constraint that required (tenant_id, name) unique for all rows
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS tenant_api_keys_unique;

-- Enforce uniqueness only among active keys (partial unique index)
-- Inactive (soft-deleted) keys are ignored, so a new key can reuse the same name
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_tenant_name_active
ON api_keys (tenant_id, name) WHERE is_active = true;
