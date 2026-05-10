-- Function DNA: Living Code That Evolves Itself In Production
-- Migration: 20260502230000_function_dna

-- ============================================================================
-- DNA Profiles — master genetic identity per function
-- ============================================================================
CREATE TABLE IF NOT EXISTS function_dna_profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id         TEXT NOT NULL,
    function_type       TEXT NOT NULL CHECK (function_type IN ('registry', 'managed')),
    tenant_id           TEXT NOT NULL,
    generation          INT NOT NULL DEFAULT 1,
    fitness_score       DOUBLE PRECISION NOT NULL DEFAULT 50.0,
    total_executions    BIGINT NOT NULL DEFAULT 0,
    total_mutations     INT NOT NULL DEFAULT 0,
    avg_latency_ms      DOUBLE PRECISION,
    p99_latency_ms      DOUBLE PRECISION,
    success_rate        DOUBLE PRECISION DEFAULT 1.0,
    error_distribution  JSONB DEFAULT '{}',
    input_patterns      JSONB DEFAULT '[]',
    bottleneck_signature JSONB DEFAULT '[]',
    dna_hash            TEXT,
    last_analyzed_at    TIMESTAMPTZ,
    evolution_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(function_id, function_type)
);

CREATE INDEX IF NOT EXISTS idx_dna_profiles_tenant ON function_dna_profiles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dna_profiles_fitness ON function_dna_profiles(fitness_score DESC);
CREATE INDEX IF NOT EXISTS idx_dna_profiles_generation ON function_dna_profiles(generation DESC);

-- ============================================================================
-- Execution Metrics — per-execution micro-data (partitioned by month)
-- ============================================================================
CREATE TABLE IF NOT EXISTS function_dna_execution_metrics (
    id                  BIGSERIAL,
    function_id         TEXT NOT NULL,
    function_type       TEXT NOT NULL CHECK (function_type IN ('registry', 'managed')),
    execution_id        TEXT,
    duration_ms         INT NOT NULL,
    memory_peak_mb      DOUBLE PRECISION,
    cpu_time_ms         INT,
    input_size_bytes    INT,
    output_size_bytes   INT,
    input_shape_hash    TEXT,
    status_code         INT,
    error_category      TEXT NOT NULL DEFAULT 'none' CHECK (error_category IN ('timeout', 'oom', 'runtime', 'logic', 'network', 'none')),
    cold_start          BOOLEAN DEFAULT FALSE,
    cache_hit           BOOLEAN DEFAULT FALSE,
    region              TEXT,
    recorded_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, recorded_at)
) PARTITION BY RANGE (recorded_at);

-- Create partitions for current + next 3 months
CREATE TABLE IF NOT EXISTS function_dna_execution_metrics_2026_05
    PARTITION OF function_dna_execution_metrics
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE IF NOT EXISTS function_dna_execution_metrics_2026_06
    PARTITION OF function_dna_execution_metrics
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE IF NOT EXISTS function_dna_execution_metrics_2026_07
    PARTITION OF function_dna_execution_metrics
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE IF NOT EXISTS function_dna_execution_metrics_2026_08
    PARTITION OF function_dna_execution_metrics
    FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

CREATE INDEX IF NOT EXISTS idx_dna_metrics_function_time
    ON function_dna_execution_metrics(function_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_dna_metrics_error
    ON function_dna_execution_metrics(function_id, error_category)
    WHERE error_category != 'none';
CREATE INDEX IF NOT EXISTS idx_dna_metrics_shape
    ON function_dna_execution_metrics(function_id, input_shape_hash);

-- ============================================================================
-- Mutations — immutable log of every evolution event
-- ============================================================================
CREATE TABLE IF NOT EXISTS function_dna_mutations (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id             TEXT NOT NULL,
    function_type           TEXT NOT NULL CHECK (function_type IN ('registry', 'managed')),
    tenant_id               TEXT NOT NULL,
    generation              INT NOT NULL,
    mutation_type           TEXT NOT NULL CHECK (mutation_type IN (
        'optimize_latency', 'reduce_memory', 'fix_error_pattern',
        'improve_reliability', 'refactor_hotpath'
    )),
    status                  TEXT NOT NULL DEFAULT 'proposed' CHECK (status IN (
        'proposed', 'accepted', 'rejected', 'deploying', 'deployed', 'rolled_back'
    )),
    trigger_reason          TEXT,
    original_code           TEXT,
    mutated_code            TEXT,
    original_hash           TEXT,
    mutated_hash            TEXT,
    diff                    TEXT,
    estimated_impact        JSONB DEFAULT '{}',
    actual_impact           JSONB,
    confidence              DOUBLE PRECISION,
    model_used              TEXT,
    analysis_window_hours   INT,
    executions_analyzed     INT,
    accepted_by             TEXT,
    accepted_at             TIMESTAMPTZ,
    deployed_at             TIMESTAMPTZ,
    rolled_back_at          TIMESTAMPTZ,
    rejected_reason         TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dna_mutations_function_gen
    ON function_dna_mutations(function_id, generation);
CREATE INDEX IF NOT EXISTS idx_dna_mutations_status
    ON function_dna_mutations(function_id, status);
CREATE INDEX IF NOT EXISTS idx_dna_mutations_tenant
    ON function_dna_mutations(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_dna_mutations_proposed
    ON function_dna_mutations(tenant_id, status)
    WHERE status = 'proposed';

-- ============================================================================
-- Insights — pre-computed enterprise analytics
-- ============================================================================
CREATE TABLE IF NOT EXISTS function_dna_insights (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                   TEXT NOT NULL,
    period_start                TIMESTAMPTZ NOT NULL,
    period_end                  TIMESTAMPTZ NOT NULL,
    total_functions_analyzed    INT NOT NULL DEFAULT 0,
    total_mutations_proposed    INT NOT NULL DEFAULT 0,
    total_mutations_accepted    INT NOT NULL DEFAULT 0,
    avg_fitness_score           DOUBLE PRECISION,
    avg_latency_improvement_pct DOUBLE PRECISION,
    total_cost_savings_usd      DOUBLE PRECISION DEFAULT 0,
    top_bottleneck_categories   JSONB DEFAULT '[]',
    evolution_leaderboard       JSONB DEFAULT '[]',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dna_insights_tenant_period
    ON function_dna_insights(tenant_id, period_start DESC);

-- ============================================================================
-- Analysis Queue — pending analysis tasks
-- ============================================================================
CREATE TABLE IF NOT EXISTS function_dna_analysis_queue (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id     TEXT NOT NULL,
    function_type   TEXT NOT NULL,
    tenant_id       TEXT NOT NULL,
    priority        INT NOT NULL DEFAULT 5,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 3,
    last_error      TEXT,
    scheduled_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dna_queue_pending
    ON function_dna_analysis_queue(priority, scheduled_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_dna_queue_function
    ON function_dna_analysis_queue(function_id, status);
