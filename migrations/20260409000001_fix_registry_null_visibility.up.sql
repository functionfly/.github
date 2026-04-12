-- Fix NULL visibility values in registry tables
-- This migration ensures all registry functions have proper visibility set

-- ============================================
-- 1. Fix NULL visibility in registry_functions
-- ============================================

-- Update any functions with NULL visibility to 'public'
UPDATE registry_functions
SET visibility = 'public',
    updated_at = NOW()
WHERE visibility IS NULL;

-- Also fix empty string visibility
UPDATE registry_functions
SET visibility = 'public',
    updated_at = NOW()
WHERE visibility = '';

-- ============================================
-- 2. Ensure default is set correctly
-- ============================================

-- If the column doesn't have a proper default, set it
ALTER TABLE registry_functions
ALTER COLUMN visibility SET DEFAULT 'public';

-- ============================================
-- 3. Verify the fix
-- ============================================

-- Count functions by visibility status
SELECT
    COALESCE(visibility, 'NULL') as visibility_status,
    COUNT(*) as count
FROM registry_functions
GROUP BY visibility
ORDER BY count DESC;
