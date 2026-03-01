-- Drop triggers
DROP TRIGGER IF EXISTS update_vulnerabilities_updated_at ON vulnerabilities;
DROP TRIGGER IF EXISTS update_security_scans_updated_at ON security_scans;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_vulnerabilities_discovered_at;
DROP INDEX IF EXISTS idx_vulnerabilities_category;
DROP INDEX IF EXISTS idx_vulnerabilities_status;
DROP INDEX IF EXISTS idx_vulnerabilities_severity;
DROP INDEX IF EXISTS idx_vulnerabilities_scan_id;
DROP INDEX IF EXISTS idx_security_scans_started_at;
DROP INDEX IF EXISTS idx_security_scans_status;
DROP INDEX IF EXISTS idx_security_scans_user_id;
DROP INDEX IF EXISTS idx_security_scans_tenant_id;

-- Drop tables
DROP TABLE IF EXISTS vulnerabilities;
DROP TABLE IF EXISTS security_scans;