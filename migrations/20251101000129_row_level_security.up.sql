-- Row Level Security (RLS) implementation for multi-tenant data isolation
-- This migration sets up RLS policies to ensure tenant data isolation

-- ============================================
-- Enable RLS on Multi-Tenant Tables
-- ============================================

-- Registry Functions - Enable RLS
ALTER TABLE registry_functions ENABLE ROW LEVEL SECURITY;

-- Registry Function Versions - Enable RLS
ALTER TABLE registry_function_versions ENABLE ROW LEVEL SECURITY;

-- Registry Function Executions - Enable RLS
ALTER TABLE registry_function_executions ENABLE ROW LEVEL SECURITY;

-- Registry Function Ratings - Enable RLS (inherits from functions)
ALTER TABLE registry_function_ratings ENABLE ROW LEVEL SECURITY;

-- Registry Function Signatures - Enable RLS
ALTER TABLE registry_function_signatures ENABLE ROW LEVEL SECURITY;

-- Registry Function Malware Scans - Enable RLS
ALTER TABLE registry_function_malware_scans ENABLE ROW LEVEL SECURITY;

-- Registry Function Approvals - Enable RLS
ALTER TABLE registry_function_approvals ENABLE ROW LEVEL SECURITY;

-- Registry Function Approval Comments - Enable RLS
ALTER TABLE registry_function_approval_comments ENABLE ROW LEVEL SECURITY;

-- Registry Function Verification Status - Enable RLS
ALTER TABLE registry_function_verification_status ENABLE ROW LEVEL SECURITY;

-- ============================================
-- Create Security Functions
-- ============================================

-- Function to get current tenant ID from session
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS uuid AS $$
BEGIN
    -- Try to get tenant_id from session variable
    BEGIN
        RETURN current_setting('app.tenant_id')::uuid;
    EXCEPTION
        WHEN others THEN
            -- Fallback: try to get from JWT claims or other context
            -- This would need to be integrated with your authentication system
            RETURN NULL;
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to get current user ID from session
CREATE OR REPLACE FUNCTION current_user_id()
RETURNS uuid AS $$
BEGIN
    BEGIN
        RETURN current_setting('app.user_id')::uuid;
    EXCEPTION
        WHEN others THEN
            RETURN NULL;
    END;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to check if current user is a tenant admin
CREATE OR REPLACE FUNCTION is_tenant_admin(user_id uuid DEFAULT current_user_id(), tenant_id uuid DEFAULT current_tenant_id())
RETURNS boolean AS $$
BEGIN
    -- Check if user has admin role for the tenant
    RETURN EXISTS (
        SELECT 1 FROM users u
        WHERE u.id = user_id
        AND u.tenant_id = tenant_id
        AND u.role IN ('admin', 'owner')
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to check if current user owns a function
CREATE OR REPLACE FUNCTION user_owns_function(function_id uuid, user_id uuid DEFAULT current_user_id())
RETURNS boolean AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.owner_user_id = user_id
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- RLS Policies for Registry Functions
-- ============================================

-- Policy: Users can view public functions from any tenant
DROP POLICY IF EXISTS registry_functions_select_public ON registry_functions;
CREATE POLICY registry_functions_select_public ON registry_functions
FOR SELECT USING (
    visibility = 'public'
);

-- Policy: Users can view functions from their own tenant
DROP POLICY IF EXISTS registry_functions_select_tenant ON registry_functions;
CREATE POLICY registry_functions_select_tenant ON registry_functions
FOR SELECT USING (
    tenant_id = current_tenant_id()
);

-- Policy: Tenant admins can view all functions in their tenant
DROP POLICY IF EXISTS registry_functions_select_tenant_admin ON registry_functions;
CREATE POLICY registry_functions_select_tenant_admin ON registry_functions
FOR SELECT USING (
    tenant_id = current_tenant_id() AND is_tenant_admin()
);

-- Policy: Users can insert functions for their tenant
DROP POLICY IF EXISTS registry_functions_insert ON registry_functions;
CREATE POLICY registry_functions_insert ON registry_functions
FOR INSERT WITH CHECK (
    tenant_id = current_tenant_id()
    AND (owner_user_id = current_user_id() OR owner_user_id IS NULL)
);

-- Policy: Function owners can update their functions
DROP POLICY IF EXISTS registry_functions_update_owner ON registry_functions;
CREATE POLICY registry_functions_update_owner ON registry_functions
FOR UPDATE USING (
    owner_user_id = current_user_id()
    OR (tenant_id = current_tenant_id() AND is_tenant_admin())
);

-- Policy: Function owners can delete their functions
DROP POLICY IF EXISTS registry_functions_delete_owner ON registry_functions;
CREATE POLICY registry_functions_delete_owner ON registry_functions
FOR DELETE USING (
    owner_user_id = current_user_id()
    OR (tenant_id = current_tenant_id() AND is_tenant_admin())
);

-- ============================================
-- RLS Policies for Registry Function Versions
-- ============================================

-- Policy: Users can view versions of public functions
DROP POLICY IF EXISTS registry_function_versions_select_public ON registry_function_versions;
CREATE POLICY registry_function_versions_select_public ON registry_function_versions
FOR SELECT USING (
    EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.visibility = 'public'
    )
);

-- Policy: Users can view versions of functions in their tenant
DROP POLICY IF EXISTS registry_function_versions_select_tenant ON registry_function_versions;
CREATE POLICY registry_function_versions_select_tenant ON registry_function_versions
FOR SELECT USING (
    EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.tenant_id = current_tenant_id()
    )
);

-- Policy: Function owners can insert new versions
DROP POLICY IF EXISTS registry_function_versions_insert_owner ON registry_function_versions;
CREATE POLICY registry_function_versions_insert_owner ON registry_function_versions
FOR INSERT WITH CHECK (
    user_owns_function(function_id)
    OR EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.tenant_id = current_tenant_id()
        AND is_tenant_admin()
    )
);

-- Policy: Function owners can update their versions
DROP POLICY IF EXISTS registry_function_versions_update_owner ON registry_function_versions;
CREATE POLICY registry_function_versions_update_owner ON registry_function_versions
FOR UPDATE USING (
    user_owns_function(function_id)
    OR EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.tenant_id = current_tenant_id()
        AND is_tenant_admin()
    )
);

-- ============================================
-- RLS Policies for Registry Function Executions
-- ============================================

-- Policy: Users can view executions of public functions
DROP POLICY IF EXISTS registry_function_executions_select_public ON registry_function_executions;
CREATE POLICY registry_function_executions_select_public ON registry_function_executions
FOR SELECT USING (
    EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.visibility = 'public'
    )
);

-- Policy: Users can view executions in their tenant
DROP POLICY IF EXISTS registry_function_executions_select_tenant ON registry_function_executions;
CREATE POLICY registry_function_executions_select_tenant ON registry_function_executions
FOR SELECT USING (
    tenant_id = current_tenant_id()
    OR EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.tenant_id = current_tenant_id()
    )
);

-- Policy: System can insert executions (this would be done by the execution engine)
DROP POLICY IF EXISTS registry_function_executions_insert_system ON registry_function_executions;
CREATE POLICY registry_function_executions_insert_system ON registry_function_executions
FOR INSERT WITH CHECK (true);  -- Allow system to insert, validation happens at application level

-- ============================================
-- RLS Policies for Registry Function Ratings
-- ============================================

-- Policy: Users can view ratings of public functions
DROP POLICY IF EXISTS registry_function_ratings_select_public ON registry_function_ratings;
CREATE POLICY registry_function_ratings_select_public ON registry_function_ratings
FOR SELECT USING (
    EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.visibility = 'public'
    )
);

-- Policy: Users can view ratings of functions in their tenant
DROP POLICY IF EXISTS registry_function_ratings_select_tenant ON registry_function_ratings;
CREATE POLICY registry_function_ratings_select_tenant ON registry_function_ratings
FOR SELECT USING (
    EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND f.tenant_id = current_tenant_id()
    )
);

-- Policy: Authenticated users can insert/update ratings
DROP POLICY IF EXISTS registry_function_ratings_upsert ON registry_function_ratings;
CREATE POLICY registry_function_ratings_upsert ON registry_function_ratings
FOR ALL USING (
    EXISTS (
        SELECT 1 FROM registry_functions f
        WHERE f.id = function_id
        AND (f.visibility = 'public' OR f.tenant_id = current_tenant_id())
    )
);

-- ============================================
-- RLS Policies for Security-Related Tables
-- ============================================

-- Policy: Only function owners and tenant admins can view signatures
DROP POLICY IF EXISTS registry_function_signatures_select_owner ON registry_function_signatures;
CREATE POLICY registry_function_signatures_select_owner ON registry_function_signatures
FOR SELECT USING (
    user_owns_function(
        (SELECT f.id FROM registry_functions f
         JOIN registry_function_versions v ON f.id = v.function_id
         WHERE v.id = function_version_id)
    )
    OR EXISTS (
        SELECT 1 FROM registry_functions f
        JOIN registry_function_versions v ON f.id = v.function_id
        WHERE v.id = function_version_id
        AND f.tenant_id = current_tenant_id()
        AND is_tenant_admin()
    )
);

-- Policy: Only function owners and tenant admins can view malware scans
DROP POLICY IF EXISTS registry_function_malware_scans_select_owner ON registry_function_malware_scans;
CREATE POLICY registry_function_malware_scans_select_owner ON registry_function_malware_scans
FOR SELECT USING (
    user_owns_function(
        (SELECT f.id FROM registry_functions f
         JOIN registry_function_versions v ON f.id = v.function_id
         WHERE v.id = function_version_id)
    )
    OR EXISTS (
        SELECT 1 FROM registry_functions f
        JOIN registry_function_versions v ON f.id = v.function_id
        WHERE v.id = function_version_id
        AND f.tenant_id = current_tenant_id()
        AND is_tenant_admin()
    )
);

-- Policy: Only reviewers can view approvals
DROP POLICY IF EXISTS registry_function_approvals_select_reviewer ON registry_function_approvals;
CREATE POLICY registry_function_approvals_select_reviewer ON registry_function_approvals
FOR SELECT USING (
    requested_by = current_user_id()
    OR assigned_to = current_user_id()
    OR user_owns_function(
        (SELECT f.id FROM registry_functions f
         JOIN registry_function_versions v ON f.id = v.function_id
         WHERE v.id = function_version_id)
    )
    OR EXISTS (
        SELECT 1 FROM registry_functions f
        JOIN registry_function_versions v ON f.id = v.function_id
        WHERE v.id = function_version_id
        AND f.tenant_id = current_tenant_id()
        AND is_tenant_admin()
    )
);

-- ============================================
-- Create Helper Functions for Application Use
-- ============================================

-- Function to set session context for a tenant/user
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id uuid, user_id uuid DEFAULT NULL)
RETURNS void AS $$
BEGIN
    -- Set session variables for RLS
    PERFORM set_config('app.tenant_id', tenant_id::text, false);
    IF user_id IS NOT NULL THEN
        PERFORM set_config('app.user_id', user_id::text, false);
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to clear session context
CREATE OR REPLACE FUNCTION clear_tenant_context()
RETURNS void AS $$
BEGIN
    -- Clear session variables
    PERFORM set_config('app.tenant_id', '', false);
    PERFORM set_config('app.user_id', '', false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to validate RLS is working
CREATE OR REPLACE FUNCTION validate_rls_setup()
RETURNS TABLE(table_name text, rls_enabled boolean, policies_count int) AS $$
BEGIN
    RETURN QUERY
    SELECT
        t.table_name::text,
        t.row_security::boolean,
        COUNT(p.policyname)::int
    FROM information_schema.tables t
    LEFT JOIN pg_policies p ON t.table_name = p.tablename
    WHERE t.table_schema = 'public'
    AND t.table_name LIKE 'registry_%'
    GROUP BY t.table_name, t.row_security
    ORDER BY t.table_name;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- Audit Triggers for RLS
-- ============================================

-- Function to audit RLS policy violations
CREATE OR REPLACE FUNCTION audit_rls_violation()
RETURNS trigger AS $$
BEGIN
    -- Log RLS policy violations (this would be triggered if someone tries to access unauthorized data)
    INSERT INTO audit_events (
        tenant_id,
        actor_user_id,
        action,
        resource_type,
        resource_id,
        details,
        created_at
    ) VALUES (
        current_tenant_id(),
        current_user_id(),
        'rls_violation',
        TG_TABLE_NAME,
        COALESCE(NEW.id::text, OLD.id::text, 'unknown'),
        jsonb_build_object(
            'operation', TG_OP,
            'table', TG_TABLE_NAME,
            'user_id', current_user_id(),
            'tenant_id', current_tenant_id()
        ),
        NOW()
    );

    -- Return NULL to prevent the operation (this is a BEFORE trigger)
    RETURN NULL;
EXCEPTION
    WHEN others THEN
        -- If audit logging fails, still prevent the operation
        RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Note: RLS policies handle the security, so we don't need violation triggers by default
-- They can be added if additional audit logging is needed

-- ============================================
-- Comments and Documentation
-- ============================================

COMMENT ON FUNCTION current_tenant_id() IS 'Returns the current tenant ID from session context';
COMMENT ON FUNCTION current_user_id() IS 'Returns the current user ID from session context';
COMMENT ON FUNCTION is_tenant_admin(uuid, uuid) IS 'Checks if a user is a tenant admin';
COMMENT ON FUNCTION user_owns_function(uuid, uuid) IS 'Checks if a user owns a specific function';
COMMENT ON FUNCTION set_tenant_context(uuid, uuid) IS 'Sets session context for tenant/user (call this at the start of requests)';
COMMENT ON FUNCTION clear_tenant_context() IS 'Clears session context (call this at the end of requests)';
COMMENT ON FUNCTION validate_rls_setup() IS 'Validates that RLS is properly configured on all tables';