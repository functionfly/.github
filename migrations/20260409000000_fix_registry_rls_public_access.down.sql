-- Revert the RLS fix for registry public access

-- ============================================
-- 1. Restore original current_tenant_id() function
-- ============================================
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS uuid AS $$
BEGIN
    -- Try to get tenant_id from session variable (old naming only)
    BEGIN
        RETURN current_setting('app.tenant_id')::uuid;
    EXCEPTION
        WHEN others THEN
            -- Fallback: return NULL
            RETURN NULL;
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- 2. Restore original current_user_id() function
-- ============================================
CREATE OR REPLACE FUNCTION current_user_id()
RETURNS uuid AS $$
BEGIN
    -- Try to get user_id from session variable (old naming only)
    BEGIN
        RETURN current_setting('app.user_id')::uuid;
    EXCEPTION
        WHEN others THEN
            -- Fallback: return NULL
            RETURN NULL;
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- 3. Drop the diagnostic function
-- ============================================
DROP FUNCTION IF EXISTS diagnose_registry_rls();

-- ============================================
-- 4. Note: We intentionally keep the RLS policies as they are correct
-- ============================================
-- The policies for public visibility are properly configured and should remain
