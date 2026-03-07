-- Migration: 000077_secrets_vault
-- Description: Rollback for zero-knowledge encrypted secrets vault
-- Drops all objects created in the up migration

-- =====================================================
-- Drop indexes for secrets_audit_log
-- =====================================================
DROP INDEX IF EXISTS idx_secrets_audit_log_tenant_time;
DROP INDEX IF EXISTS idx_secrets_audit_log_tenant;
DROP INDEX IF EXISTS idx_secrets_audit_log_timestamp;
DROP INDEX IF EXISTS idx_secrets_audit_log_secret;

-- =====================================================
-- Drop indexes for secret_access_tokens
-- =====================================================
DROP INDEX IF EXISTS idx_secret_access_tokens_active;
DROP INDEX IF EXISTS idx_secret_access_tokens_hash;
DROP INDEX IF EXISTS idx_secret_access_tokens_expires;
DROP INDEX IF EXISTS idx_secret_access_tokens_secret;

-- =====================================================
-- Drop indexes for secrets_vault
-- =====================================================
DROP INDEX IF EXISTS idx_secrets_vault_active;
DROP INDEX IF EXISTS idx_secrets_vault_tenant_name;
DROP INDEX IF EXISTS idx_secrets_vault_tenant;

-- =====================================================
-- Drop tables in reverse order (respecting foreign key dependencies)
-- =====================================================

-- Drop audit log table first (references secrets_vault)
DROP TABLE IF EXISTS secrets_audit_log;

-- Drop access tokens table (references secrets_vault)
DROP TABLE IF EXISTS secret_access_tokens;

-- Drop main secrets vault table last
DROP TABLE IF EXISTS secrets_vault;
