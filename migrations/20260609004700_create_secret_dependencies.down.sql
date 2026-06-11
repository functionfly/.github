DROP INDEX IF EXISTS idx_secret_dependencies_unique;
DROP INDEX IF EXISTS idx_secret_dependencies_dependent_type;
DROP INDEX IF EXISTS idx_secret_dependencies_dependent_id;
DROP INDEX IF EXISTS idx_secret_dependencies_tenant_id;
DROP INDEX IF EXISTS idx_secret_dependencies_secret_id;
DROP TABLE IF EXISTS secret_dependencies;