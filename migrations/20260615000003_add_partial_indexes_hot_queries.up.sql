-- Migration: 20260615000003_add_partial_indexes_hot_queries
-- Description: Add targeted partial indexes for high-frequency filtered query patterns
-- Addresses: signals queries, consciousness_insights, agent_messages, registry_functions

-- ============================================
-- 1. signals: Add tenant + created_at composite index
-- ============================================
CREATE INDEX IF NOT EXISTS idx_signals_tenant_created
    ON signals(tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_signals_tenant_type
    ON signals(tenant_id, signal_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_signals_tenant_connector
    ON signals(tenant_id, connector_slug, created_at DESC)
    WHERE connector_slug IS NOT NULL;

-- ============================================
-- 2. consciousness_insights: Partial indexes for status filtering
-- ============================================
CREATE INDEX IF NOT EXISTS idx_consciousness_insights_pending
    ON consciousness_insights(tenant_id, created_at DESC)
    WHERE status = 'pending';

CREATE INDEX IF EXISTS idx_consciousness_insights_analyzing
    ON consciousness_insights(tenant_id, created_at DESC)
    WHERE status = 'analyzing';

CREATE INDEX IF NOT EXISTS idx_consciousness_insights_delivered
    ON consciousness_insights(tenant_id, delivered_at DESC)
    WHERE status = 'delivered';

-- ============================================
-- 3. agent_messages: Partial indexes for pending/to_agent queries
-- ============================================
CREATE INDEX IF NOT EXISTS idx_agent_messages_to_agent_pending
    ON agent_messages(to_agent_id, created_at ASC)
    WHERE delivered_at IS NULL AND read_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_agent_messages_to_agent_read
    ON agent_messages(to_agent_id, read_at DESC)
    WHERE read_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_agent_messages_from_agent
    ON agent_messages(from_agent_id, created_at DESC);

-- ============================================
-- 4. registry_functions: Partial indexes for approved/published functions
-- ============================================
CREATE INDEX IF NOT EXISTS idx_registry_functions_approved
    ON registry_functions(tenant_id, created_at DESC)
    WHERE status = 'approved' AND published_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_registry_functions_latest
    ON registry_functions(tenant_id, created_at DESC)
    WHERE latest_version IS NOT NULL;

-- ============================================
-- 5. factory_runs: Partial indexes for running/completed states
-- ============================================
CREATE INDEX IF NOT EXISTS idx_factory_runs_running
    ON factory_runs(tenant_id, created_at DESC)
    WHERE status = 'running';

CREATE INDEX IF NOT EXISTS idx_factory_runs_failed
    ON factory_runs(tenant_id, finished_at DESC)
    WHERE status = 'failed';

-- ============================================
-- 6. agent_identities: Index for active agent lookups
-- ============================================
CREATE INDEX IF NOT EXISTS idx_agent_identities_tenant_active
    ON agent_identities(tenant_id, created_at DESC)
    WHERE deactivated_at IS NULL;

-- ============================================
-- 7. webhook_configs: Index for active tenant webhooks
-- ============================================
CREATE INDEX IF NOT EXISTS idx_webhook_configs_tenant_active
    ON webhook_configs(tenant_id, created_at DESC)
    WHERE is_active = TRUE;

COMMENT ON INDEX idx_signals_tenant_created IS 'Supports brain_repository.go signals queries filtering by tenant and time';
COMMENT ON INDEX idx_consciousness_insights_pending IS 'Supports consciousness insight polling for pending items';
COMMENT ON INDEX idx_agent_messages_to_agent_pending IS 'Supports message delivery polling with efficient pending filter';
