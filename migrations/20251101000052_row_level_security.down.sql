-- Drop Row Level Security policies and functions

-- Drop RLS policies
DROP POLICY IF EXISTS registry_functions_select_public ON registry_functions;
DROP POLICY IF EXISTS registry_functions_select_tenant ON registry_functions;
DROP POLICY IF EXISTS registry_functions_select_tenant_admin ON registry_functions;
DROP POLICY IF EXISTS registry_functions_insert ON registry_functions;
DROP POLICY IF EXISTS registry_functions_update_owner ON registry_functions;
DROP POLICY IF EXISTS registry_functions_delete_owner ON registry_functions;

DROP POLICY IF EXISTS registry_function_versions_select_public ON registry_function_versions;
DROP POLICY IF EXISTS registry_function_versions_select_tenant ON registry_function_versions;
DROP POLICY IF EXISTS registry_function_versions_insert_owner ON registry_function_versions;
DROP POLICY IF EXISTS registry_function_versions_update_owner ON registry_function_versions;

DROP POLICY IF EXISTS registry_function_executions_select_public ON registry_function_executions;
DROP POLICY IF EXISTS registry_function_executions_select_tenant ON registry_function_executions;
DROP POLICY IF EXISTS registry_function_executions_insert_system ON registry_function_executions;

DROP POLICY IF EXISTS registry_function_ratings_select_public ON registry_function_ratings;
DROP POLICY IF EXISTS registry_function_ratings_select_tenant ON registry_function_ratings;
DROP POLICY IF EXISTS registry_function_ratings_upsert ON registry_function_ratings;

DROP POLICY IF EXISTS registry_function_signatures_select_owner ON registry_function_signatures;
DROP POLICY IF EXISTS registry_function_malware_scans_select_owner ON registry_function_malware_scans;
DROP POLICY IF EXISTS registry_function_approvals_select_reviewer ON registry_function_approvals;

-- Disable RLS on tables
ALTER TABLE registry_functions DISABLE ROW LEVEL SECURITY;
ALTER TABLE registry_function_versions DISABLE ROW LEVEL SECURITY;
ALTER TABLE registry_function_executions DISABLE ROW LEVEL SECURITY;
ALTER TABLE registry_function_ratings DISABLE ROW LEVEL SECURITY;
ALTER TABLE registry_function_signatures DISABLE ROW LEVEL SECURITY;
ALTER TABLE registry_function_malware_scans DISABLE ROW LEVEL SECURITY;
ALTER TABLE registry_function_approvals DISABLE ROW LEVEL SECURITY;
ALTER TABLE registry_function_approval_comments DISABLE ROW LEVEL SECURITY;
ALTER TABLE registry_function_verification_status DISABLE ROW LEVEL SECURITY;

-- Drop helper functions
DROP FUNCTION IF EXISTS current_tenant_id();
DROP FUNCTION IF EXISTS current_user_id();
DROP FUNCTION IF EXISTS is_tenant_admin(uuid, uuid);
DROP FUNCTION IF EXISTS user_owns_function(uuid, uuid);
DROP FUNCTION IF EXISTS set_tenant_context(uuid, uuid);
DROP FUNCTION IF EXISTS clear_tenant_context();
DROP FUNCTION IF EXISTS validate_rls_setup();
DROP FUNCTION IF EXISTS audit_rls_violation();