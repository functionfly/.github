-- Drop all GitHub integration tables in reverse dependency order
DROP TABLE IF EXISTS github_import_templates;
DROP TABLE IF EXISTS github_sync_logs;
DROP TABLE IF EXISTS github_webhooks;
DROP TABLE IF EXISTS github_imports;
DROP TABLE IF EXISTS github_repos;
DROP TABLE IF EXISTS github_connections;
