-- Rollback IP allowlist tables
-- Migration: 20260307000002_create_ip_allowlist_tables.down.sql

-- Drop indexes first
DROP INDEX IF EXISTS idx_ip_allowlist_entries_allowlist;
DROP INDEX IF EXISTS idx_ip_allowlist_tenant;

-- Drop tables (order matters due to foreign key constraint)
DROP TABLE IF EXISTS ip_allowlist_entries;
DROP TABLE IF EXISTS ip_allowlists;
