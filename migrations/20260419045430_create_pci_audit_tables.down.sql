-- Rollback PCI DSS Audit Logging Tables
-- WARNING: This removes all PCI compliance audit data

-- Drop helper function
DROP FUNCTION IF EXISTS pci_purge_expired_audit_events();

-- Drop tables (in correct order to avoid dependency issues)
DROP TABLE IF EXISTS pci_environment_controls;
DROP TABLE IF EXISTS pci_cardholder_data_access_logs;
DROP TABLE IF EXISTS pci_key_access_logs;
DROP TABLE IF EXISTS pci_encryption_keys;
DROP TABLE IF EXISTS pci_audit_events;
