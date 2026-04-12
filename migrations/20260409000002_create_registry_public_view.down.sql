-- Revert the registry public view

-- ============================================
-- 1. Drop the functions
-- ============================================
DROP FUNCTION IF EXISTS list_public_registry_functions(VARCHAR, VARCHAR, INTEGER, INTEGER);
DROP FUNCTION IF EXISTS count_public_registry_functions(VARCHAR, VARCHAR);

-- ============================================
-- 2. Drop the view
-- ============================================
DROP VIEW IF EXISTS v_registry_functions_public CASCADE;
