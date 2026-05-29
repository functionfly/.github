-- Migration: Fix is_under_legal_hold function for proper legal hold checking
-- Created: 2026-05-28
-- Bug fix: Function was being called with tenant_id instead of timestamp,
-- and didn't check tenant_id scope properly.

-- ============================================================
-- Updated Function to check if an entry is under legal hold
-- ============================================================
-- Old signature: is_under_legal_hold(p_table_name, p_date_from, p_date_to)
-- New signature: is_under_legal_hold(p_table_name, p_tenant_id, p_entry_timestamp)
--
-- This function checks:
-- 1. Table scope (NULL = all tables, or specific table name)
-- 2. Tenant scope (NULL = all tenants, or specific tenant_id)
-- 3. Entry timestamp falls within the hold's date range
-- 4. Hold is active (status = 'active' and not expired)

CREATE OR REPLACE FUNCTION is_under_legal_hold(
    p_table_name VARCHAR(100),
    p_tenant_id UUID,
    p_entry_timestamp TIMESTAMP WITH TIME ZONE
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM legal_holds
        WHERE status = 'active'
          AND (expires_at IS NULL OR expires_at > NOW())
          AND (table_name IS NULL OR table_name = p_table_name)
          AND (tenant_id IS NULL OR tenant_id = p_tenant_id)
          AND (
              -- Entry timestamp must fall within hold date range
              (hold_date_from IS NULL OR hold_date_from <= p_entry_timestamp)
              AND (hold_date_to IS NULL OR hold_date_to >= p_entry_timestamp)
          )
    );
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- Comments for documentation
-- ============================================================
COMMENT ON FUNCTION is_under_legal_hold IS
    'Checks if a specific entry (identified by table, tenant, timestamp) is under an active legal hold. Returns true if deletion should be blocked.';
