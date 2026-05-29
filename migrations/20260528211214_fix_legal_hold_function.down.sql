-- Migration: Revert is_under_legal_hold function to original signature
-- Created: 2026-05-28

CREATE OR REPLACE FUNCTION is_under_legal_hold(
    p_table_name VARCHAR(100),
    p_date_from TIMESTAMP WITH TIME ZONE,
    p_date_to TIMESTAMP WITH TIME ZONE
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM legal_holds
        WHERE status = 'active'
          AND (table_name IS NULL OR table_name = p_table_name)
          AND (expires_at IS NULL OR expires_at > NOW())
          AND (
              -- Overlapping date range check
              (hold_date_from IS NULL OR hold_date_from <= p_date_to)
              AND (hold_date_to IS NULL OR hold_date_to >= p_date_from)
          )
    );
END;
$$ LANGUAGE plpgsql;
