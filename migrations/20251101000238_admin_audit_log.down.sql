-- Drop admin audit log table
DROP INDEX IF EXISTS idx_admin_audit_log_user_action;
DROP INDEX IF EXISTS idx_admin_audit_log_action;
DROP INDEX IF EXISTS idx_admin_audit_log_created_at;
DROP INDEX IF EXISTS idx_admin_audit_log_resource;
DROP INDEX IF EXISTS idx_admin_audit_log_user_id;
DROP TABLE IF EXISTS admin_audit_log;
