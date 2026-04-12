-- Down migration for privacy tables
DROP TABLE IF EXISTS privacy_audit_log;
DROP TABLE IF EXISTS data_deletion_requests;
DROP TABLE IF EXISTS data_export_requests;
DROP TABLE IF EXISTS privacy_consent_records;
DROP TABLE IF EXISTS global_privacy_settings;
DROP TABLE IF EXISTS privacy_settings;
