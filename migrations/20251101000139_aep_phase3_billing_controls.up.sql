-- AEP Phase 3: Agent Billing Controls
-- Migration: 0003_aep_phase3_billing_controls

CREATE TABLE IF NOT EXISTS agent_billing_controls (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id                TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    spend_cap_monthly_usd   DECIMAL(10,2),
    spend_cap_daily_usd     DECIMAL(10,2),
    credit_balance_usd      DECIMAL(10,2) NOT NULL DEFAULT 0,
    billing_mode            TEXT NOT NULL DEFAULT 'per_agent',  -- per_agent | per_tenant | per_team
    team_id                 UUID,
    alert_thresholds        DECIMAL[] NOT NULL DEFAULT '{0.5, 0.8, 0.95}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_billing_controls_agent_id ON agent_billing_controls(agent_id);
