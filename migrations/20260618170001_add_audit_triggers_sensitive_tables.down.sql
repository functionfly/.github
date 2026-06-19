-- Migration: Rollback additional audit triggers
-- Created: 2026-06-18

DROP TRIGGER IF EXISTS audit_functions_trigger ON functions;
DROP TRIGGER IF EXISTS audit_api_keys_trigger ON api_keys;
DROP TRIGGER IF EXISTS audit_secrets_vault_trigger ON secrets_vault;
DROP TRIGGER IF EXISTS audit_legal_holds_trigger ON legal_holds;
DROP TRIGGER IF EXISTS audit_invoices_trigger ON invoices;
DROP TRIGGER IF EXISTS audit_cost_allocation_entries_trigger ON cost_allocation_entries;
DROP TRIGGER IF EXISTS audit_agent_identities_trigger ON agent_identities;
