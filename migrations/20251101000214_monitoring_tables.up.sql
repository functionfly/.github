-- Monitoring and observability tables for enhanced Supabase monitoring
-- Includes performance metrics, alerts, and real-time monitoring data

-- =============================================
-- PERFORMANCE METRICS TABLE
-- =============================================
CREATE TABLE IF NOT EXISTS performance_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_type VARCHAR(50) NOT NULL, -- 'response_time', 'error_rate', 'throughput', 'health_score'
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    backend_id UUID REFERENCES backends(id) ON DELETE CASCADE,
    value DECIMAL(10,2) NOT NULL,
    unit VARCHAR(20) NOT NULL, -- 'ms', 'percent', 'requests_per_second', 'score'
    labels JSONB DEFAULT '{}', -- Additional metadata like region, provider, etc.
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================
-- ALERTS AND INCIDENTS TABLE
-- =============================================
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_type VARCHAR(50) NOT NULL, -- 'health_degraded', 'backend_down', 'high_error_rate', 'circuit_open'
    severity VARCHAR(20) NOT NULL DEFAULT 'info', -- 'info', 'warning', 'error', 'critical'
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    backend_id UUID REFERENCES backends(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'acknowledged', 'resolved'
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID REFERENCES users(id),
    metadata JSONB DEFAULT '{}', -- Additional alert-specific data
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================
-- SYSTEM HEALTH CHECKS TABLE
-- =============================================
CREATE TABLE IF NOT EXISTS system_health_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    check_type VARCHAR(50) NOT NULL, -- 'database', 'api', 'external_service', 'disk_space', 'memory'
    component_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'unknown', -- 'healthy', 'degraded', 'unhealthy', 'unknown'
    response_time_ms INTEGER,
    message TEXT,
    metadata JSONB DEFAULT '{}',
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================
-- REAL-TIME MONITORING EVENTS TABLE
-- =============================================
CREATE TABLE IF NOT EXISTS monitoring_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL, -- 'request_completed', 'backend_failover', 'circuit_breaker_transition'
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    app_id UUID REFERENCES apps(id) ON DELETE CASCADE,
    backend_id UUID REFERENCES backends(id) ON DELETE CASCADE,
    request_id VARCHAR(255),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    data JSONB DEFAULT '{}', -- Event-specific payload
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================
-- DASHBOARD CONFIGURATIONS TABLE
-- =============================================
CREATE TABLE IF NOT EXISTS dashboard_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL for tenant-wide configs
    config_type VARCHAR(50) NOT NULL, -- 'metric_panel', 'alert_rule', 'chart_config'
    name VARCHAR(100) NOT NULL,
    config JSONB NOT NULL, -- Configuration data specific to the type
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, user_id, config_type, name) -- Prevent duplicate configs
);

-- =============================================
-- INDEXES FOR PERFORMANCE
-- =============================================

-- Performance metrics indexes
CREATE INDEX IF NOT EXISTS idx_performance_metrics_type_timestamp ON performance_metrics(metric_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_tenant ON performance_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_app ON performance_metrics(app_id);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_backend ON performance_metrics(backend_id);

-- Alerts indexes
CREATE INDEX IF NOT EXISTS idx_alerts_type_status ON alerts(alert_type, status);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_tenant ON alerts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_alerts_created ON alerts(created_at DESC);

-- System health checks indexes
CREATE INDEX IF NOT EXISTS idx_system_health_checks_type ON system_health_checks(check_type);
CREATE INDEX IF NOT EXISTS idx_system_health_checks_status ON system_health_checks(status);
CREATE INDEX IF NOT EXISTS idx_system_health_checks_checked ON system_health_checks(checked_at DESC);

-- Monitoring events indexes
CREATE INDEX IF NOT EXISTS idx_monitoring_events_type ON monitoring_events(event_type);
CREATE INDEX IF NOT EXISTS idx_monitoring_events_tenant ON monitoring_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_monitoring_events_timestamp ON monitoring_events(timestamp DESC);

-- Dashboard configs indexes
CREATE INDEX IF NOT EXISTS idx_dashboard_configs_tenant ON dashboard_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dashboard_configs_user ON dashboard_configs(user_id);

-- =============================================
-- REALTIME PUBLICATIONS FOR MONITORING
-- =============================================

-- Enable realtime for monitoring tables
-- Note: These would be configured in Supabase dashboard or via separate migration