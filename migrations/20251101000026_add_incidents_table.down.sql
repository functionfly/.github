-- Drop incidents table and indexes
DROP INDEX IF EXISTS idx_incidents_resolved_at;
DROP INDEX IF EXISTS idx_incidents_created_at;
DROP INDEX IF EXISTS idx_incidents_severity;
DROP INDEX IF EXISTS idx_incidents_status;
DROP TABLE IF EXISTS incidents;