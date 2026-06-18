-- Rollback: 20260615000003_add_partial_indexes_hot_queries

DROP INDEX IF EXISTS idx_signals_tenant_created;
DROP INDEX IF EXISTS idx_signals_tenant_type;
DROP INDEX IF EXISTS idx_signals_tenant_connector;
DROP INDEX IF EXISTS idx_consciousness_insights_pending;
DROP INDEX IF EXISTS idx_consciousness_insights_analyzing;
DROP INDEX IF EXISTS idx_consciousness_insights_delivered;
DROP INDEX IF EXISTS idx_agent_messages_to_agent_pending;
DROP INDEX IF EXISTS idx_agent_messages_to_agent_read;
DROP INDEX IF EXISTS idx_agent_messages_from_agent;
DROP INDEX IF EXISTS idx_registry_functions_approved;
DROP INDEX IF EXISTS idx_registry_functions_latest;
DROP INDEX IF EXISTS idx_factory_runs_running;
DROP INDEX IF EXISTS idx_factory_runs_failed;
DROP INDEX IF EXISTS idx_agent_identities_tenant_active;
DROP INDEX IF EXISTS idx_webhook_configs_tenant_active;
