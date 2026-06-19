-- Migration: Add audit triggers for sensitive tables
-- Created: 2026-06-18
-- Purpose: Extend audit coverage to more sensitive tables that were missing triggers
--
-- Tables already audited: users, apps, backends, deployments, subscriptions
-- Tables being added: functions, api_keys, secrets_vault, legal_holds, invoices, cost_allocation_entries, agent_identities
-- Note: tenant_payment_methods and webhook_configs tables don't exist in this schema

-- ============================================================
-- Additional tables to audit
-- ============================================================

-- Functions - core business logic table
DROP TRIGGER IF EXISTS audit_functions_trigger ON functions;
CREATE TRIGGER audit_functions_trigger
    AFTER INSERT OR UPDATE OR DELETE ON functions
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- API keys - sensitive credentials
DROP TRIGGER IF EXISTS audit_api_keys_trigger ON api_keys;
CREATE TRIGGER audit_api_keys_trigger
    AFTER INSERT OR UPDATE OR DELETE ON api_keys
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Secrets vault - sensitive credential storage
DROP TRIGGER IF EXISTS audit_secrets_vault_trigger ON secrets_vault;
CREATE TRIGGER audit_secrets_vault_trigger
    AFTER INSERT OR UPDATE OR DELETE ON secrets_vault
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Legal holds - compliance table, important to track changes
DROP TRIGGER IF EXISTS audit_legal_holds_trigger ON legal_holds;
CREATE TRIGGER audit_legal_holds_trigger
    AFTER INSERT OR UPDATE OR DELETE ON legal_holds
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Invoices - financial records
DROP TRIGGER IF EXISTS audit_invoices_trigger ON invoices;
CREATE TRIGGER audit_invoices_trigger
    AFTER INSERT OR UPDATE OR DELETE ON invoices
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Cost allocation entries - billing data
DROP TRIGGER IF EXISTS audit_cost_allocation_entries_trigger ON cost_allocation_entries;
CREATE TRIGGER audit_cost_allocation_entries_trigger
    AFTER INSERT OR UPDATE OR DELETE ON cost_allocation_entries
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Agent identities - important for AI agent tracking
DROP TRIGGER IF EXISTS audit_agent_identities_trigger ON agent_identities;
CREATE TRIGGER audit_agent_identities_trigger
    AFTER INSERT OR UPDATE OR DELETE ON agent_identities
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- ============================================================
-- Note: The audit_trigger_function() handles these tables via
-- the existing CASE statement for tenant_id extraction.
-- If a table has a 'tenant_id' column, it will be captured.
-- For tables without tenant_id, NULL will be used.
-- ============================================================

COMMENT ON TRIGGER audit_functions_trigger ON functions IS
    'Audits all changes to function definitions';
COMMENT ON TRIGGER audit_api_keys_trigger ON api_keys IS
    'Audits all changes to API keys - important for security';
COMMENT ON TRIGGER audit_secrets_vault_trigger ON secrets_vault IS
    'Audits all changes to secrets vault entries';
COMMENT ON TRIGGER audit_legal_holds_trigger ON legal_holds IS
    'Audits all changes to legal hold records - compliance requirement';
COMMENT ON TRIGGER audit_invoices_trigger ON invoices IS
    'Audits all changes to invoices - financial compliance';
COMMENT ON TRIGGER audit_cost_allocation_entries_trigger ON cost_allocation_entries IS
    'Audits all changes to cost allocation - billing compliance';
COMMENT ON TRIGGER audit_agent_identities_trigger ON agent_identities IS
    'Audits all changes to agent identities';
