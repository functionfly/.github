-- Migration: Remove retention audit and legal holds tables (rollback)
-- Created: 2026-04-19

-- Drop triggers
DROP TRIGGER IF EXISTS trg_legal_holds_updated_at ON legal_holds;

-- Drop functions
DROP FUNCTION IF EXISTS update_legal_hold_updated_at();
DROP FUNCTION IF EXISTS is_under_legal_hold(VARCHAR(100), TIMESTAMP WITH TIME ZONE, TIMESTAMP WITH TIME ZONE);

-- Drop tables (cascade to remove dependent objects)
DROP TABLE IF EXISTS legal_holds CASCADE;
DROP TABLE IF EXISTS retention_audit_log CASCADE;
