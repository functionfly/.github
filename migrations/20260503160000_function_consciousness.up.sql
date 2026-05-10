-- Function Consciousness: Predictive awareness layer
-- Creates tables for insights, awareness scores, preferences, and delivery log

-- Consciousness insights (the main table)
CREATE TABLE IF NOT EXISTS consciousness_insights (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Insight classification
    category        VARCHAR(50) NOT NULL,
    severity        VARCHAR(20) NOT NULL,
    priority        INT NOT NULL DEFAULT 0,
    
    -- Content
    title           VARCHAR(500) NOT NULL,
    message         TEXT NOT NULL,
    summary         VARCHAR(200),
    
    -- Context
    function_id     UUID,
    graph_id        UUID,
    agent_id        UUID,
    related_function_ids UUID[] DEFAULT '{}',
    
    -- Data payload
    insight_data    JSONB NOT NULL DEFAULT '{}',
    action_type     VARCHAR(50),
    action_data     JSONB DEFAULT '{}',
    action_preview  JSONB DEFAULT '{}',
    
    -- Trajectory (for predictive insights)
    trajectory      VARCHAR(20),
    projected_days  INT,
    confidence      NUMERIC(5,4),
    
    -- Lifecycle
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    dismissed_at    TIMESTAMPTZ,
    applied_at      TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    superseded_by   UUID REFERENCES consciousness_insights(id),
    
    -- Delivery tracking
    channels_sent   TEXT[] DEFAULT '{}',
    read_at         TIMESTAMPTZ,
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consciousness_insights_tenant_status ON consciousness_insights(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_consciousness_insights_tenant_category ON consciousness_insights(tenant_id, category);
CREATE INDEX IF NOT EXISTS idx_consciousness_insights_created ON consciousness_insights(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_consciousness_insights_function ON consciousness_insights(function_id) WHERE function_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_consciousness_insights_priority ON consciousness_insights(tenant_id, priority DESC, created_at DESC) WHERE status = 'active';

-- System Awareness Score (one per tenant, updated periodically)
CREATE TABLE IF NOT EXISTS system_awareness_scores (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Overall score (0-100)
    overall_score   NUMERIC(5,2) NOT NULL DEFAULT 0,
    
    -- Component scores (0-100 each)
    health_score    NUMERIC(5,2) NOT NULL DEFAULT 0,
    efficiency_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    scalability_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    reliability_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    optimization_score NUMERIC(5,2) NOT NULL DEFAULT 0,
    
    -- Metadata
    functions_analyzed  INT NOT NULL DEFAULT 0,
    graphs_analyzed     INT NOT NULL DEFAULT 0,
    agents_analyzed     INT NOT NULL DEFAULT 0,
    active_insights     INT NOT NULL DEFAULT 0,
    critical_insights   INT NOT NULL DEFAULT 0,
    
    -- Trend
    previous_score  NUMERIC(5,2),
    trend           VARCHAR(20),
    
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_system_awareness_scores_tenant ON system_awareness_scores(tenant_id);

-- Consciousness notification preferences (one per tenant)
CREATE TABLE IF NOT EXISTS consciousness_preferences (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Channel preferences
    email_enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    slack_enabled       BOOLEAN NOT NULL DEFAULT FALSE,
    slack_webhook_url   VARCHAR(1000),
    inapp_enabled       BOOLEAN NOT NULL DEFAULT TRUE,
    webhook_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    webhook_url         VARCHAR(1000),
    webhook_secret      VARCHAR(256),
    
    -- Frequency preferences
    digest_frequency    VARCHAR(20) NOT NULL DEFAULT 'daily',
    quiet_hours_start   TIME,
    quiet_hours_end     TIME,
    timezone            VARCHAR(50) DEFAULT 'UTC',
    
    -- Category preferences
    enabled_categories  TEXT[] DEFAULT ARRAY['traffic','cost','redundancy','health','marketplace','scaling'],
    
    -- Severity filter
    min_notify_severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    
    -- Autonomous actions (Agent Enterprise only)
    auto_apply_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
    auto_apply_categories TEXT[] DEFAULT '{}',
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_consciousness_preferences_tenant ON consciousness_preferences(tenant_id);

-- Insight delivery log (for analytics and dedup)
CREATE TABLE IF NOT EXISTS consciousness_delivery_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    insight_id      UUID NOT NULL REFERENCES consciousness_insights(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel         VARCHAR(20) NOT NULL,
    status          VARCHAR(20) NOT NULL,
    error_message   TEXT,
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consciousness_delivery_log_tenant ON consciousness_delivery_log(tenant_id, sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_consciousness_delivery_log_insight ON consciousness_delivery_log(insight_id);
