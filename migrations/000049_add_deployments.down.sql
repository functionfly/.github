-- Remove deployment tracking tables

-- Drop indexes first
DROP INDEX IF EXISTS idx_deployments_app_id;
DROP INDEX IF EXISTS idx_deployments_provider;
DROP INDEX IF EXISTS idx_deployments_status;
DROP INDEX IF EXISTS idx_deployments_created_at;
DROP INDEX IF EXISTS idx_deployment_artifacts_app_id;

-- Drop tables
DROP TABLE IF EXISTS deployment_artifacts;
DROP TABLE IF EXISTS deployments;