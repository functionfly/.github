-- Fix Row Level Security for public registry access
-- Problem: When RLS is enabled on registry_functions, the API's database user may not be
-- able to see public functions due to missing context or policy configuration.
--
-- Solution: Update RLS policies to ensure public visibility works correctly

-- ============================================
-- 1. Fix current_tenant_id() function to handle both naming conventions
-- ============================================

-- Drop and recreate with proper handling for both old and new naming
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS uuid AS $$
BEGIN
    -- Try newer naming first (app.current_tenant_id)
    BEGIN
        RETURN NULLIF(current_setting('app.current_tenant_id', true), '')::uuid;
    EXCEPTION
        WHEN others THEN
            -- Try older naming (app.tenant_id)
            BEGIN
                RETURN NULLIF(current_setting('app.tenant_id', true), '')::uuid;
            EXCEPTION
                WHEN others THEN
                    RETURN NULL;
            END;
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- 2. Fix current_user_id() function similarly
-- ============================================
CREATE OR REPLACE FUNCTION current_user_id()
RETURNS uuid AS $$
BEGIN
    -- Try newer naming first
    BEGIN
        RETURN NULLIF(current_setting('app.current_user_id', true), '')::uuid;
    EXCEPTION
        WHEN others THEN
            -- Try older naming
            BEGIN
                RETURN NULLIF(current_setting('app.user_id', true), '')::uuid;
            EXCEPTION
                WHEN others THEN
                    RETURN NULL;
            END;
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- 3. Fix RLS Policies for Registry Functions
-- ============================================

-- Drop existing public policy to recreate it cleanly
DROP POLICY IF EXISTS registry_functions_select_public ON registry_functions;

-- Create a permissive policy that allows ANY user to see public functions
-- The visibility = 'public' check is independent of tenant context
CREATE POLICY registry_functions_select_public ON registry_functions
FOR SELECT
USING (visibility = 'public');

-- Ensure RLS is enabled
ALTER TABLE registry_functions ENABLE ROW LEVEL SECURITY;

-- ============================================
-- 4. Fix RLS Policies for Registry Function Versions
-- ============================================

-- Drop and recreate version policy
DROP POLICY IF EXISTS registry_function_versions_select_public ON registry_function_versions;

CREATE POLICY registry_function_versions_select_public ON registry_function_versions
FOR SELECT
USING (
    EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.visibility = 'public'
    )
);

ALTER TABLE registry_function_versions ENABLE ROW LEVEL SECURITY;

-- ============================================
-- 5. Fix RLS Policies for Registry Ratings
-- ============================================

DROP POLICY IF EXISTS registry_function_ratings_select_public ON registry_function_ratings;

CREATE POLICY registry_function_ratings_select_public ON registry_function_ratings
FOR SELECT
USING (
    EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.visibility = 'public'
    )
);

ALTER TABLE registry_function_ratings ENABLE ROW LEVEL SECURITY;

-- ============================================
-- 6. Ensure visibility is set correctly on existing functions
-- ============================================

-- Update any functions with NULL or empty visibility to 'public'
UPDATE registry_functions
SET visibility = 'public',
    updated_at = NOW()
WHERE visibility IS NULL OR visibility = '';

-- ============================================
-- 7. Create helper function to diagnose RLS issues
-- ============================================

CREATE OR REPLACE FUNCTION diagnose_registry_rls()
RETURNS TABLE (
    table_name text,
    rls_enabled boolean,
    policy_count bigint,
    public_function_count bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        t.tablename::text,
        c.relrowsecurity as rls_enabled,
        (SELECT COUNT(*) FROM pg_policies WHERE tablename = t.tablename) as policy_count,
        (SELECT COUNT(*) FROM registry_functions WHERE visibility = 'public') as public_function_count
    FROM pg_tables t
    JOIN pg_class c ON c.relname = t.tablename
    WHERE t.schemaname = 'public'
    AND t.tablename LIKE 'registry_%';
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- 8. Verify the fix
-- ============================================

-- Output diagnostic info
SELECT * FROM diagnose_registry_rls();
