-- Rollback platform maintenance mode tables

DROP TABLE IF EXISTS maintenance_audit_log CASCADE;
DROP TABLE IF EXISTS maintenance_page_templates CASCADE;
DROP TABLE IF EXISTS platform_maintenance CASCADE;
