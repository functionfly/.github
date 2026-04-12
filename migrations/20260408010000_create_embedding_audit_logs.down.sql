-- Migration: Drop embedding audit logs table
-- Reverse of: 20260408000000_create_embedding_audit_logs.up.sql

-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS embedding_daily_costs;

-- Drop the summary view
DROP VIEW IF EXISTS embedding_audit_summary;

-- Drop the cleanup function
DROP FUNCTION IF EXISTS cleanup_old_embedding_audit_logs();

-- Drop the main table (indexes will be dropped automatically)
DROP TABLE IF EXISTS embedding_audit_logs;

-- Note: We keep the pgcrypto extension as it may be used by other tables
