-- Migration to create cost_allocation_entries table for detailed cost tracking
-- This enables fine-grained chargebacks and cost transparency

CREATE TABLE IF NOT EXISTS cost_allocation_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    function_id UUID NOT NULL,
    function_name VARCHAR(255) NOT NULL,
    function_author VARCHAR(255) NOT NULL,
    execution_id UUID NOT NULL,
    execution_outcome VARCHAR(50) NOT NULL,
    cached BOOLEAN NOT NULL DEFAULT FALSE,

    -- Resource usage
    duration_ms BIGINT NOT NULL DEFAULT 0,
    cpu_time_ms BIGINT NOT NULL DEFAULT 0,
    memory_used_mb BIGINT NOT NULL DEFAULT 0,
    wall_time_ms BIGINT NOT NULL DEFAULT 0,

    -- Cost breakdown (in cents for precision)
    execution_cost_cents BIGINT NOT NULL DEFAULT 0,
    compute_cost_cents BIGINT NOT NULL DEFAULT 0,
    platform_fee_cents BIGINT NOT NULL DEFAULT 0,
    data_transfer_cents BIGINT NOT NULL DEFAULT 0,
    total_cost_cents BIGINT NOT NULL DEFAULT 0,

    -- Metadata
    region VARCHAR(50) NOT NULL DEFAULT 'unknown',
    timestamp TIMESTAMPTZ NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    tags JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT positive_duration CHECK (duration_ms >= 0),
    CONSTRAINT positive_cpu_time CHECK (cpu_time_ms >= 0),
    CONSTRAINT positive_costs CHECK (total_cost_cents >= 0)
);

-- Indexes for efficient querying
CREATE INDEX idx_cost_allocation_tenant_timestamp ON cost_allocation_entries(tenant_id, timestamp DESC);
CREATE INDEX idx_cost_allocation_function ON cost_allocation_entries(function_id, timestamp DESC);
CREATE INDEX idx_cost_allocation_period ON cost_allocation_entries(period_start, period_end);
CREATE INDEX idx_cost_allocation_region ON cost_allocation_entries(region);
CREATE INDEX idx_cost_allocation_timestamp ON cost_allocation_entries(timestamp DESC);
CREATE INDEX idx_cost_allocation_created_at ON cost_allocation_entries(created_at DESC);

-- Partial indexes for common queries
CREATE INDEX idx_cost_allocation_tenant_success ON cost_allocation_entries(tenant_id, timestamp DESC)
    WHERE execution_outcome = 'success';
CREATE INDEX idx_cost_allocation_tenant_errors ON cost_allocation_entries(tenant_id, timestamp DESC)
    WHERE execution_outcome = 'error';

-- Composite index for date range queries
-- Note: Partial index with NOW() is not allowed (functions must be IMMUTABLE)
-- This regular composite index efficiently supports date range queries
CREATE INDEX idx_cost_allocation_date_range ON cost_allocation_entries(tenant_id, timestamp);

-- GIN index for tags and metadata
CREATE INDEX idx_cost_allocation_tags ON cost_allocation_entries USING GIN(tags);
CREATE INDEX idx_cost_allocation_metadata ON cost_allocation_entries USING GIN(metadata);

-- Table for cost allocation report history
CREATE TABLE IF NOT EXISTS cost_allocation_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    tenant_count INTEGER NOT NULL DEFAULT 0,
    function_count INTEGER NOT NULL DEFAULT 0,
    total_executions BIGINT NOT NULL DEFAULT 0,
    total_cost_cents BIGINT NOT NULL DEFAULT 0,
    report_data JSONB NOT NULL DEFAULT '{}',
    generated_by UUID,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cost_allocation_reports_period ON cost_allocation_reports(period_start, period_end);
CREATE INDEX idx_cost_allocation_reports_generated_at ON cost_allocation_reports(generated_at DESC);

-- Add comment for documentation
COMMENT ON TABLE cost_allocation_entries IS 'Detailed cost allocation records for function executions, enabling fine-grained chargebacks and cost transparency';
COMMENT ON TABLE cost_allocation_reports IS 'History of generated cost allocation reports for audit and reference';
