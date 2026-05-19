-- Drop studio_tasks table
DROP INDEX IF EXISTS idx_studio_tasks_tenant_status;
DROP INDEX IF EXISTS idx_studio_tasks_created_at;
DROP INDEX IF EXISTS idx_studio_tasks_created_by;
DROP INDEX IF EXISTS idx_studio_tasks_assignee;
DROP INDEX IF EXISTS idx_studio_tasks_status;
DROP INDEX IF EXISTS idx_studio_tasks_tenant;
DROP TABLE IF EXISTS studio_tasks;