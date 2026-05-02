-- Migration: Drop trust audit log table
-- Reverses: 20260412175601_create_trust_audit_log

DROP TABLE IF EXISTS trust_audit_logs;