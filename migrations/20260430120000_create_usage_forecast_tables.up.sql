-- Usage Forecasting and Alerting Tables
-- Creates tables for predictive analytics and proactive spend/usage alerts

CREATE TABLE IF NOT EXISTS usage_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    threshold_value DECIMAL(15,2) NOT NULL,
    threshold_operator VARCHAR(20) NOT NULL DEFAULT 'gte',
    period_type VARCHAR(50) NOT NULL DEFAULT 'billing_period',
    notification_channels VARCHAR(50)[] DEFAULT ARRAY['email', 'in_app'],
    is_enabled BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    trigger_count INTEGER DEFAULT 0,
    cooldown_minutes INTEGER DEFAULT 60,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_alerts_tenant ON usage_alerts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_usage_alerts_enabled ON usage_alerts(is_enabled) WHERE is_enabled = true;
CREATE INDEX IF NOT EXISTS idx_usage_alerts_type ON usage_alerts(alert_type);

CREATE TABLE IF NOT EXISTS usage_alert_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID NOT NULL REFERENCES usage_alerts(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    triggered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    triggered_value DECIMAL(15,2) NOT NULL,
    threshold_value DECIMAL(15,2) NOT NULL,
    message TEXT,
    metadata JSONB,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by UUID REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_alert_history_tenant ON usage_alert_history(tenant_id);
CREATE INDEX IF NOT EXISTS idx_alert_history_alert ON usage_alert_history(alert_id);
CREATE INDEX IF NOT EXISTS idx_alert_history_triggered ON usage_alert_history(triggered_at);
CREATE INDEX IF NOT EXISTS idx_alert_history_unack ON usage_alert_history(acknowledged_at) WHERE acknowledged_at IS NULL;

CREATE TABLE IF NOT EXISTS spend_caps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    cap_amount_cents INTEGER NOT NULL,
    warning_thresholds INTEGER[] DEFAULT ARRAY[50, 75, 90],
    current_spend_cents INTEGER DEFAULT 0,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    action_on_cap VARCHAR(50) DEFAULT 'notify_only',
    is_hard_cap BOOLEAN DEFAULT false,
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, period_start)
);

CREATE INDEX IF NOT EXISTS idx_spend_caps_tenant ON spend_caps(tenant_id);
CREATE INDEX IF NOT EXISTS idx_spend_caps_period ON spend_caps(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_spend_caps_enabled ON spend_caps(is_enabled) WHERE is_enabled = true;

CREATE TABLE IF NOT EXISTS usage_forecasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    forecast_type VARCHAR(50) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    current_value DECIMAL(15,2) NOT NULL,
    predicted_value DECIMAL(15,2) NOT NULL,
    lower_bound DECIMAL(15,2) NOT NULL,
    upper_bound DECIMAL(15,2) NOT NULL,
    confidence DECIMAL(5,4) NOT NULL,
    method_used VARCHAR(50) NOT NULL,
    growth_rate DECIMAL(10,6),
    days_of_history INTEGER NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_usage_forecasts_tenant ON usage_forecasts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_usage_forecasts_type ON usage_forecasts(forecast_type);
CREATE INDEX IF NOT EXISTS idx_usage_forecasts_created ON usage_forecasts(created_at);
CREATE INDEX IF NOT EXISTS idx_usage_forecasts_tenant_type_created ON usage_forecasts(tenant_id, forecast_type, created_at);

CREATE TABLE IF NOT EXISTS usage_trends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    period_analyzed VARCHAR(20) NOT NULL,
    avg_daily_usage DECIMAL(15,4),
    peak_daily_usage DECIMAL(15,4),
    min_daily_usage DECIMAL(15,4),
    trend_direction VARCHAR(20),
    trend_percent_change DECIMAL(10,4),
    seasonality_score DECIMAL(5,4),
    volatility_score DECIMAL(10,4),
    anomaly_count INTEGER DEFAULT 0,
    forecast_accuracy DECIMAL(5,4),
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, event_type, period_analyzed)
);

CREATE INDEX IF NOT EXISTS idx_usage_trends_tenant ON usage_trends(tenant_id);
CREATE INDEX IF NOT EXISTS idx_usage_trends_calculated ON usage_trends(calculated_at);

CREATE OR REPLACE FUNCTION update_usage_alerts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS usage_alerts_updated_at ON usage_alerts;
CREATE TRIGGER usage_alerts_updated_at
    BEFORE UPDATE ON usage_alerts
    FOR EACH ROW EXECUTE FUNCTION update_usage_alerts_updated_at();

CREATE OR REPLACE FUNCTION update_spend_caps_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS spend_caps_updated_at ON spend_caps;
CREATE TRIGGER spend_caps_updated_at
    BEFORE UPDATE ON spend_caps
    FOR EACH ROW EXECUTE FUNCTION update_spend_caps_updated_at();
