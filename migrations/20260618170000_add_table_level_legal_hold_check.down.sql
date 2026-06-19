-- Migration: Rollback table-level legal hold check function
-- Created: 2026-06-18

-- Drop the 1-arg overload
DROP FUNCTION IF EXISTS is_under_legal_hold(VARCHAR);

-- Note: The 3-arg function is kept as-is since it was added in a prior migration
