-- Rollback: dispute_automation_tables

DROP TABLE IF EXISTS dispute_customer_notifications;
DROP TABLE IF EXISTS dispute_evidence_cache;
DROP TABLE IF EXISTS dispute_automation_config;
DROP TABLE IF EXISTS dispute_automation_log;

DROP INDEX IF EXISTS idx_payment_disputes_evidence_due_by;
DROP INDEX IF EXISTS idx_payment_disputes_status_pending;
