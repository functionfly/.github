-- StateFabric: Composable Durable State for Stateless Functions
-- Core state tables for globally addressable, deterministic state

-- 1. States table - State containers bound to function identities
CREATE TABLE IF NOT EXISTS states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Naming and addressing
    name VARCHAR(255) NOT NULL,
    full_path VARCHAR(512) NOT NULL, -- "acme/cart"
    function_id UUID REFERENCES registry_functions(id), -- Optional bound function

    -- State configuration
    storage_type VARCHAR(50) DEFAULT 'keyvalue' CHECK (storage_type IN ('keyvalue', 'document', 'timeseries', 'graph')),

    -- Retention
    ttl_days INTEGER DEFAULT 0, -- 0 = forever
    max_size_mb INTEGER DEFAULT 100,

    -- Versioning
    current_version INTEGER DEFAULT 1,
    is_versioned BOOLEAN DEFAULT true,

    -- Permissions
    is_public BOOLEAN DEFAULT false,
    allow_cross_tenant BOOLEAN DEFAULT false,

    -- Metadata
    description TEXT,
    tags JSONB DEFAULT '[]'::jsonb,

    -- Billing and usage
    storage_used_mb BIGINT DEFAULT 0,
    write_ops_month BIGINT DEFAULT 0,
    read_ops_month BIGINT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_states_tenant_path UNIQUE (tenant_id, full_path)
);

-- Indexes for states
CREATE INDEX IF NOT EXISTS idx_states_tenant_id ON states(tenant_id);
CREATE INDEX IF NOT EXISTS idx_states_function_id ON states(function_id);
CREATE INDEX IF NOT EXISTS idx_states_full_path ON states(full_path);
CREATE INDEX IF NOT EXISTS idx_states_created_at ON states(created_at);


-- 2. State values table - Key-value entries within state
CREATE TABLE IF NOT EXISTS state_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,

    -- Key supports hierarchical keys like "user/123/profile"
    key VARCHAR(1024) NOT NULL,

    -- Value stored as JSON for flexibility
    value JSONB NOT NULL,

    -- Versioning
    version INTEGER NOT NULL DEFAULT 1,
    previous_value JSONB,

    -- Content addressing (for deduplication)
    content_hash VARCHAR(64),

    -- TTL
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Metadata
    created_by VARCHAR(255), -- function_id or user_id
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_state_values_state_key_version UNIQUE (state_id, key, version)
);

-- Indexes for state_values
CREATE INDEX IF NOT EXISTS idx_state_values_state_id ON state_values(state_id);
CREATE INDEX IF NOT EXISTS idx_state_values_key ON state_values(key);
CREATE INDEX IF NOT EXISTS idx_state_values_content_hash ON state_values(content_hash);
CREATE INDEX IF NOT EXISTS idx_state_values_expires_at ON state_values(expires_at) WHERE expires_at IS NOT NULL;
-- CREATE INDEX IF NOT EXISTS idx_state_values_latest ON state_values(state_id, key) WHERE version = (SELECT MAX(version) FROM state_values sv WHERE sv.state_id = state_values.state_id AND sv.key = state_values.key); -- Subqueries not allowed in index predicates


-- 3. State events table - Immutable event log for replay
CREATE TABLE IF NOT EXISTS state_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,

    -- Event types: "set" | "delete" | "snapshot" | "restore" | "merge"
    event_type VARCHAR(50) NOT NULL,

    -- Key affected (null for state-level events)
    key VARCHAR(1024),

    -- Event data
    previous_value JSONB,
    new_value JSONB,

    -- Causality
    causation_id UUID, -- Link to triggering event
    correlation_id VARCHAR(255), -- For distributed tracing

    -- Source
    source_type VARCHAR(50), -- "function" | "user" | "system" | "trigger"
    source_id VARCHAR(255), -- function_id or user_id

    -- Determinism proof (for replay verification)
    input_hash VARCHAR(64),
    output_hash VARCHAR(64),
    deterministic BOOLEAN DEFAULT false,

    -- Sequence (for ordering)
    sequence_num BIGINT NOT NULL,

    -- Timestamp
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for state_events
CREATE INDEX IF NOT EXISTS idx_state_events_state_id ON state_events(state_id);
CREATE INDEX IF NOT EXISTS idx_state_events_event_type ON state_events(event_type);
CREATE INDEX IF NOT EXISTS idx_state_events_key ON state_events(key);
CREATE INDEX IF NOT EXISTS idx_state_events_timestamp ON state_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_state_events_correlation_id ON state_events(correlation_id);
CREATE INDEX IF NOT EXISTS idx_state_events_sequence ON state_events(state_id, sequence_num);


-- 4. State snapshots table - Versioned state snapshots
CREATE TABLE IF NOT EXISTS state_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,

    -- Snapshot identification
    snapshot_version INTEGER NOT NULL,
    label VARCHAR(255),

    -- Content
    state_data JSONB NOT NULL,
    state_size_bytes BIGINT,

    -- Coverage
    key_count INTEGER,
    first_sequence BIGINT,
    last_sequence BIGINT,

    -- Determinism
    root_event_id UUID, -- First event in snapshot

    -- Compression
    is_compressed BOOLEAN DEFAULT false,
    compression_algo VARCHAR(20), -- "lz4", "zstd", ""

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_state_snapshots_state_version UNIQUE (state_id, snapshot_version)
);

-- Indexes for state_snapshots
CREATE INDEX IF NOT EXISTS idx_state_snapshots_state_id ON state_snapshots(state_id);
CREATE INDEX IF NOT EXISTS idx_state_snapshots_created_at ON state_snapshots(created_at);


-- 5. State permissions table - Access control for state
CREATE TABLE IF NOT EXISTS state_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,

    -- Principal
    principal_type VARCHAR(50) NOT NULL, -- "user" | "team" | "function" | "tenant"
    principal_id UUID,

    -- Permissions
    can_read BOOLEAN DEFAULT false,
    can_write BOOLEAN DEFAULT false,
    can_delete BOOLEAN DEFAULT false,
    can_admin BOOLEAN DEFAULT false,
    can_trigger BOOLEAN DEFAULT false, -- For function triggers

    -- Constraints
    ip_whitelist JSONB DEFAULT '[]'::jsonb,
    time_restrictions JSONB,
    rate_limit INTEGER, -- Requests per minute

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for state_permissions
CREATE INDEX IF NOT EXISTS idx_state_permissions_state_id ON state_permissions(state_id);
CREATE INDEX IF NOT EXISTS idx_state_permissions_principal ON state_permissions(principal_type, principal_id);


-- 6. Function trigger types enum
DO $$ BEGIN
    CREATE TYPE trigger_type AS ENUM ('on_write', 'on_read', 'on_delete', 'on_condition');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 7. State triggers table - Automatic function invocation on state changes
CREATE TABLE IF NOT EXISTS state_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Source
    source_state_id UUID REFERENCES states(id) ON DELETE CASCADE,

    -- Trigger configuration
    trigger_type trigger_type NOT NULL,
    key_pattern VARCHAR(512), -- Glob pattern for keys

    -- Condition (for advanced triggers)
    condition JSONB,

    -- Target function
    target_function_id UUID REFERENCES registry_functions(id),
    target_function VARCHAR(255), -- "org/function:version"

    -- Payload
    include_previous BOOLEAN DEFAULT false,
    include_new BOOLEAN DEFAULT true,

    -- Rate limiting
    max_invocations_per_minute INTEGER DEFAULT 60,

    -- Status
    is_active BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for state_triggers
CREATE INDEX IF NOT EXISTS idx_state_triggers_tenant_id ON state_triggers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_state_triggers_source_state_id ON state_triggers(source_state_id);
CREATE INDEX IF NOT EXISTS idx_state_triggers_target_function_id ON state_triggers(target_function_id);
CREATE INDEX IF NOT EXISTS idx_state_triggers_is_active ON state_triggers(is_active);


-- 8. Agent memory types enum
DO $$ BEGIN
    CREATE TYPE memory_type AS ENUM ('working', 'longterm', 'context', 'episodic');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 9. Agent memories table - AI agent memory with embeddings
CREATE TABLE IF NOT EXISTS agent_memories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    agent_id VARCHAR(255) NOT NULL,

    -- Memory type
    memory_type memory_type NOT NULL,

    -- Content
    content TEXT,
    structured_data JSONB,

    -- Embedding (BYTEA when pgvector not installed; use vector(1536) if you have CREATE EXTENSION vector)
    embedding BYTEA,

    -- Metadata
    importance_score REAL DEFAULT 0.5, -- 0.0-1.0 for retention
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMP WITH TIME ZONE,

    -- Retention
    ttl_days INTEGER DEFAULT 0, -- 0 = forever
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Causality
    source_event_id UUID,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for agent_memories
CREATE INDEX IF NOT EXISTS idx_agent_memories_tenant_id ON agent_memories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_agent_memories_agent_id ON agent_memories(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_memories_memory_type ON agent_memories(memory_type);
CREATE INDEX IF NOT EXISTS idx_agent_memories_expires_at ON agent_memories(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_agent_memories_importance ON agent_memories(importance_score);


-- 10. Agent memory index table - Track index configuration per agent
CREATE TABLE IF NOT EXISTS agent_memory_indexes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    agent_id VARCHAR(255) NOT NULL,

    -- Index configuration
    memory_type memory_type NOT NULL,
    dimension INTEGER DEFAULT 1536,
    similarity_metric VARCHAR(20) DEFAULT 'cosine',

    -- Index stats
    memory_count INTEGER DEFAULT 0,
    last_indexed_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_memory_indexes_agent ON agent_memory_indexes(tenant_id, agent_id, memory_type);


-- 11. State usage metrics table - For billing and analytics
CREATE TABLE IF NOT EXISTS state_usage_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    state_id UUID REFERENCES states(id) ON DELETE CASCADE,

    -- Metric type
    metric_type VARCHAR(50) NOT NULL, -- "storage" | "write_ops" | "read_ops"

    -- Value
    value BIGINT NOT NULL,
    unit VARCHAR(20), -- "bytes" | "ops" | "mb"

    -- Time period
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for state_usage_metrics
CREATE INDEX IF NOT EXISTS idx_state_usage_metrics_tenant_id ON state_usage_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_state_usage_metrics_state_id ON state_usage_metrics(state_id);
CREATE INDEX IF NOT EXISTS idx_state_usage_metrics_period ON state_usage_metrics(period_start, period_end);
