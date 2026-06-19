-- Migration: Add table-level legal hold check function overload
-- Created: 2026-06-18
-- Purpose: The archive-service needs to check if a table has ANY active legal hold
-- before deletion. This 1-arg overload checks table-level holds without requiring
-- tenant_id or timestamp (which the archive-service doesn't have at check time).

-- ============================================================
-- Function to check if a table has any active legal hold
-- ============================================================
-- This checks:
-- 1. Table scope (NULL = all tables, or specific table name)
-- 2. Hold is active (status = 'active' and not expired)
-- Note: This is a table-level check without date range or tenant scope

CREATE OR REPLACE FUNCTION is_under_legal_hold(
    p_table_name VARCHAR(100)
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM legal_holds
        WHERE status = 'active'
          AND (expires_at IS NULL OR expires_at > NOW())
          AND (table_name IS NULL OR table_name = p_table_name)
    );
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- Keep the 3-arg function for entry-level checks (tenant + timestamp)
-- ============================================================
-- This is used by the data retention scheduler to check if specific
-- data entries fall within a legal hold's date range.

-- Already exists from 20260528211214_fix_legal_hold_function.up.sql
-- Signature: is_under_legal_hold(p_table_name, p_tenant_id, p_entry_timestamp)

-- ============================================================
-- Comments for documentation
-- ============================================================
COMMENT ON FUNCTION is_under_legal_hold(VARCHAR) IS
    'Checks if a table has any active legal hold (no date/tenant scope). Used for table-level deletion checks.';
