-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS internal_opportunities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL DEFAULT 'internal_rd_lab',
    tags JSONB DEFAULT '[]',
    demand_score DECIMAL(5,2) DEFAULT 0,
    quality_score DECIMAL(5,2) DEFAULT 0,
    complexity INT NOT NULL DEFAULT 1,
    validated BOOLEAN NOT NULL DEFAULT true,
    status TEXT NOT NULL DEFAULT 'pending',
    estimated_value_usd DECIMAL(12,2) DEFAULT 0,
    priority TEXT DEFAULT 'medium',
    confidential BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_internal_opportunities_status ON internal_opportunities(status);
CREATE INDEX IF NOT EXISTS idx_internal_opportunities_priority ON internal_opportunities(priority);
CREATE INDEX IF NOT EXISTS idx_internal_opportunities_confidential ON internal_opportunities(confidential);
CREATE INDEX IF NOT EXISTS idx_internal_opportunities_value ON internal_opportunities(estimated_value_usd);

CREATE TABLE IF NOT EXISTS internal_functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL,
    tags JSONB DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'prototype',
    source_opportunity_id UUID,
    generated_code TEXT,
    estimated_value DECIMAL(12,2) DEFAULT 0,
    time_to_value TEXT,
    privacy_level TEXT DEFAULT 'internal_team_only',
    competitive_edge JSONB,
    rd_phase TEXT DEFAULT 'prototype',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_internal_functions_status ON internal_functions(status);
CREATE INDEX IF NOT EXISTS idx_internal_functions_privacy ON internal_functions(privacy_level);

CREATE TABLE IF NOT EXISTS rd_lab_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status TEXT NOT NULL,
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    duration BIGINT,
    ideas_scouted INT DEFAULT 0,
    ideas_funded INT DEFAULT 0,
    prototypes INT DEFAULT 0,
    total_value_tracked DECIMAL(12,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS stealth_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status TEXT NOT NULL,
    mode TEXT,
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    ops_executed INT DEFAULT 0,
    functions_built INT DEFAULT 0,
    value_generated DECIMAL(12,2) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS swarm_metrics_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_at TIMESTAMP DEFAULT NOW(),
    date DATE,
    total_runs INT DEFAULT 0,
    successful_runs INT DEFAULT 0,
    failed_runs INT DEFAULT 0,
    functions_generated INT DEFAULT 0,
    functions_published INT DEFAULT 0,
    average_quality_score DECIMAL(5,2) DEFAULT 0,
    new_opportunities BIGINT DEFAULT 0,
    conversion_rate DECIMAL(5,2) DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_swarm_metrics_snapshots_date ON swarm_metrics_snapshots(date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS swarm_metrics_snapshots;
DROP TABLE IF EXISTS stealth_runs;
DROP TABLE IF EXISTS rd_lab_runs;
DROP TABLE IF EXISTS internal_functions;
DROP TABLE IF EXISTS internal_opportunities;
-- +goose StatementEnd