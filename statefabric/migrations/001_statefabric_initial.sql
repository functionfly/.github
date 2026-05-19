-- StateFabric PostgreSQL Schema (Metadata Only)
-- Stores pointers/indexes to object storage, not blob data

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- States Table (metadata index)
-- ============================================
CREATE TABLE IF NOT EXISTS states (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    path VARCHAR(500) NOT NULL,
    full_path VARCHAR(1000) NOT NULL,
    current_version BIGINT NOT NULL DEFAULT 0,
    storage_type VARCHAR(50) NOT NULL DEFAULT 'keyvalue',
    state_hash VARCHAR(128),
    size_bytes BIGINT DEFAULT 0,
    key_count INTEGER DEFAULT 0,
    deterministic BOOLEAN NOT NULL DEFAULT FALSE,
    agent_id UUID,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_tenant_path UNIQUE (tenant_id, path)
);

CREATE INDEX IF NOT EXISTS idx_states_tenant_id ON states(tenant_id);
CREATE INDEX IF NOT EXISTS idx_states_full_path ON states(full_path);
CREATE INDEX IF NOT EXISTS idx_states_agent_id ON states(agent_id);
CREATE INDEX IF NOT EXISTS idx_states_created_at ON states(created_at);

-- ============================================
-- Event Metadata Table (pointers to object storage)
-- ============================================
CREATE TABLE IF NOT EXISTS event_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    key VARCHAR(500),
    correlation_id VARCHAR(255) NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    source_id VARCHAR(255) NOT NULL,
    deterministic BOOLEAN NOT NULL DEFAULT FALSE,
    sequence_num BIGINT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    storage_key VARCHAR(1000) NOT NULL,
    input_hash VARCHAR(128),
    output_hash VARCHAR(128),

    CONSTRAINT unique_state_sequence UNIQUE (state_id, sequence_num)
);

CREATE INDEX IF NOT EXISTS idx_event_metadata_state_id ON event_metadata(state_id);
CREATE INDEX IF NOT EXISTS idx_event_metadata_sequence ON event_metadata(state_id, sequence_num);
CREATE INDEX IF NOT EXISTS idx_event_metadata_correlation_id ON event_metadata(correlation_id);
CREATE INDEX IF NOT EXISTS idx_event_metadata_timestamp ON event_metadata(timestamp);
CREATE INDEX IF NOT EXISTS idx_event_metadata_source ON event_metadata(source_type, source_id);

-- ============================================
-- Snapshot Metadata Table (pointers to object storage)
-- ============================================
CREATE TABLE IF NOT EXISTS snapshot_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,
    snapshot_version BIGINT NOT NULL,
    label VARCHAR(500),
    key_count INTEGER DEFAULT 0,
    size_bytes BIGINT DEFAULT 0,
    first_sequence BIGINT NOT NULL,
    last_sequence BIGINT NOT NULL,
    root_event_id UUID NOT NULL,
    is_compressed BOOLEAN DEFAULT FALSE,
    compression_algo VARCHAR(50),
    checksum VARCHAR(128),
    storage_key VARCHAR(1000) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_state_snapshot_version UNIQUE (state_id, snapshot_version)
);

CREATE INDEX IF NOT EXISTS idx_snapshot_metadata_state_id ON snapshot_metadata(state_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_metadata_version ON snapshot_metadata(state_id, snapshot_version);
CREATE INDEX IF NOT EXISTS idx_snapshot_metadata_created_at ON snapshot_metadata(created_at);

-- ============================================
-- Agent Configuration Table
-- ============================================
CREATE TABLE IF NOT EXISTS agent_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    agent_id UUID NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_agent_config UNIQUE (tenant_id, agent_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_configs_tenant_id ON agent_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_agent_configs_agent_id ON agent_configs(agent_id);

-- ============================================
-- Billing Counters Table
-- ============================================
CREATE TABLE IF NOT EXISTS billing_counters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    counter_type VARCHAR(100) NOT NULL,
    period VARCHAR(7) NOT NULL, -- YYYY-MM format
    count BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_tenant_period_type UNIQUE (tenant_id, counter_type, period)
);

CREATE INDEX IF NOT EXISTS idx_billing_counters_tenant ON billing_counters(tenant_id);
CREATE INDEX IF NOT EXISTS idx_billing_counters_period ON billing_counters(period);

-- ============================================
-- Execution Hashes Table (for replay verification)
-- ============================================
CREATE TABLE IF NOT EXISTS execution_hashes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,
    execution_id VARCHAR(255) NOT NULL,
    input_hash VARCHAR(128) NOT NULL,
    output_hash VARCHAR(128) NOT NULL,
    verified_at TIMESTAMPTZ,
    verification_result VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_execution UNIQUE (state_id, execution_id)
);

CREATE INDEX IF NOT EXISTS idx_execution_hashes_state_id ON execution_hashes(state_id);
CREATE INDEX IF NOT EXISTS idx_execution_hashes_execution_id ON execution_hashes(execution_id);

-- ============================================
-- Function: Update updated_at trigger
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_states_updated_at
    BEFORE UPDATE ON states
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_agent_configs_updated_at
    BEFORE UPDATE ON agent_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Comments for documentation
-- ============================================
COMMENT ON TABLE states IS 'State metadata index - stores pointers to state data in object storage';
COMMENT ON TABLE event_metadata IS 'Event metadata index - stores pointers to event data in object storage';
COMMENT ON TABLE snapshot_metadata IS 'Snapshot metadata index - stores pointers to snapshot data in object storage';
COMMENT ON TABLE agent_configs IS 'Agent configuration metadata';
COMMENT ON TABLE billing_counters IS 'Usage tracking for billing';
COMMENT ON TABLE execution_hashes IS 'Execution hashes for deterministic replay verification';
