-- Migration: 000253_security_alert_rules.down.sql
-- Description: Drop security_alert_rules table
-- Created: 2026-03-22

-- Drop trigger and function first
DROP TRIGGER IF EXISTS trigger_security_alert_rules_updated_at ON security_alert_rules;
DROP FUNCTION IF EXISTS update_security_alert_rules_updated_at();

-- Drop the table
DROP TABLE IF EXISTS security_alert_rules;
