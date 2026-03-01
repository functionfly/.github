-- State Fabrics (dashboard): fabric containers with stores, pipelines, events, snapshots, replays

-- 1. State fabrics - top-level container per tenant
CREATE TABLE IF NOT EXISTS state_fabrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('online', 'offline', 'degraded', 'pending', 'suspended')),
    type VARCHAR(50) NOT NULL DEFAULT 'custom' CHECK (type IN ('session', 'catalog', 'cache', 'workflow', 'custom')),
    settings JSONB NOT NULL DEFAULT '{"autoSnapshot":false,"snapshotIntervalMinutes":60,"retentionDays":30,"enableReplication":false,"regions":[],"conflictResolution":"last-write-wins"}',
    throughput BIGINT NOT NULL DEFAULT 0,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    suspended_at TIMESTAMP WITH TIME ZONE,
    suspend_reason TEXT
);

CREATE INDEX IF NOT EXISTS idx_state_fabrics_tenant_id ON state_fabrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_state_fabrics_status ON state_fabrics(status);
CREATE INDEX IF NOT EXISTS idx_state_fabrics_created_at ON state_fabrics(created_at);

-- 2. State fabric stores
CREATE TABLE IF NOT EXISTS state_fabric_stores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fabric_id UUID NOT NULL REFERENCES state_fabrics(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('memory', 'persistent', 'cache', 'queue')),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'error')),
    size BIGINT NOT NULL DEFAULT 0,
    max_size BIGINT NOT NULL DEFAULT 0,
    region VARCHAR(100) NOT NULL DEFAULT 'default',
    provider VARCHAR(100) NOT NULL DEFAULT 'local',
    throughput BIGINT NOT NULL DEFAULT 0,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_state_fabric_stores_fabric_id ON state_fabric_stores(fabric_id);

-- 3. State fabric pipelines
CREATE TABLE IF NOT EXISTS state_fabric_pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fabric_id UUID NOT NULL REFERENCES state_fabrics(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('active', 'paused', 'error', 'draft')),
    steps JSONB NOT NULL DEFAULT '[]',
    input_schema JSONB,
    output_schema JSONB,
    throughput BIGINT NOT NULL DEFAULT 0,
    error_rate REAL NOT NULL DEFAULT 0,
    last_executed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_state_fabric_pipelines_fabric_id ON state_fabric_pipelines(fabric_id);

-- 4. State fabric events (event log)
CREATE TABLE IF NOT EXISTS state_fabric_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fabric_id UUID NOT NULL REFERENCES state_fabrics(id) ON DELETE CASCADE,
    store_id UUID REFERENCES state_fabric_stores(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('create', 'update', 'delete', 'snapshot', 'sync')),
    payload JSONB NOT NULL DEFAULT '{}',
    sequence_number BIGINT NOT NULL,
    correlation_id VARCHAR(255),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_state_fabric_events_fabric_id ON state_fabric_events(fabric_id);
CREATE INDEX IF NOT EXISTS idx_state_fabric_events_store_id ON state_fabric_events(store_id);
CREATE INDEX IF NOT EXISTS idx_state_fabric_events_timestamp ON state_fabric_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_state_fabric_events_sequence ON state_fabric_events(fabric_id, sequence_number);

-- 5. State fabric snapshots
CREATE TABLE IF NOT EXISTS state_fabric_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fabric_id UUID NOT NULL REFERENCES state_fabrics(id) ON DELETE CASCADE,
    store_id UUID REFERENCES state_fabric_stores(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    state_data JSONB NOT NULL DEFAULT '{}',
    event_count INTEGER NOT NULL DEFAULT 0,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_state_fabric_snapshots_fabric_id ON state_fabric_snapshots(fabric_id);
CREATE INDEX IF NOT EXISTS idx_state_fabric_snapshots_store_id ON state_fabric_snapshots(store_id);

-- 6. State fabric replays
CREATE TABLE IF NOT EXISTS state_fabric_replays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fabric_id UUID NOT NULL REFERENCES state_fabrics(id) ON DELETE CASCADE,
    snapshot_id UUID REFERENCES state_fabric_snapshots(id) ON DELETE SET NULL,
    start_event_id UUID REFERENCES state_fabric_events(id) ON DELETE SET NULL,
    end_event_id UUID REFERENCES state_fabric_events(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    progress INTEGER NOT NULL DEFAULT 0,
    events_replayed BIGINT NOT NULL DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_state_fabric_replays_fabric_id ON state_fabric_replays(fabric_id);

-- 7. Pipeline executions (for execute endpoint)
CREATE TABLE IF NOT EXISTS state_fabric_pipeline_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES state_fabric_pipelines(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    input_data JSONB,
    output_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_state_fabric_pipeline_executions_pipeline_id ON state_fabric_pipeline_executions(pipeline_id);
