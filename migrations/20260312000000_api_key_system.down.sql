-- Migration: 20260312000000_api_key_system
-- Description: Rollback API Key System migrations
-- Created: 2026-03-12
-- Author: FunctionFly

-- Drop indexes first
DROP INDEX IF EXISTS idx_api_key_environments_env_id;
DROP INDEX IF EXISTS idx_api_key_environments_key_id;
DROP INDEX IF EXISTS idx_api_key_permissions_resource;
DROP INDEX IF EXISTS idx_api_key_permissions_key_id;
DROP INDEX IF EXISTS idx_api_key_rotations_rotated_at;
DROP INDEX IF EXISTS idx_api_key_rotations_key_id;
DROP INDEX IF EXISTS idx_api_keys_type;
DROP INDEX IF EXISTS idx_api_keys_expires;
DROP INDEX IF EXISTS idx_api_keys_active;
DROP INDEX IF EXISTS idx_api_keys_hash;
DROP INDEX IF EXISTS idx_api_keys_prefix;
DROP INDEX IF EXISTS idx_api_keys_tenant_name;
DROP INDEX IF EXISTS idx_api_keys_tenant;

-- Drop tables in reverse order (respecting foreign key dependencies)
DROP TABLE IF EXISTS api_key_environments;
DROP TABLE IF EXISTS api_key_permissions;
DROP TABLE IF EXISTS api_key_rotations;
DROP TABLE IF EXISTS api_keys;
