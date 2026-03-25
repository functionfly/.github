-- Drop IP allowlist table
DROP INDEX IF EXISTS idx_ip_allowlist_created_by;
DROP INDEX IF EXISTS idx_ip_allowlist_active;
DROP INDEX IF EXISTS idx_ip_allowlist_cidr;
DROP TABLE IF EXISTS ip_allowlist;
