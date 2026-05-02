-- Drop email workflow tables

DROP INDEX IF EXISTS idx_email_workflow_executions_scheduled_at;
DROP INDEX IF EXISTS idx_email_workflow_executions_status;
DROP INDEX IF EXISTS idx_email_workflow_executions_workflow_id;
DROP INDEX IF EXISTS idx_email_workflow_executions_tenant_id;

DROP TABLE IF EXISTS email_workflow_executions;

DROP INDEX IF EXISTS idx_email_workflow_configs_active;
DROP INDEX IF EXISTS idx_email_workflow_configs_bundle_slug;
DROP INDEX IF EXISTS idx_email_workflow_configs_tenant_id;

DROP TABLE IF EXISTS email_workflow_configs;
