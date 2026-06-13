-- FOM: Function Outcome Model Database
-- Migration: 001_fom_initial.sql
-- Date: 2026-06-12

-- ============================================================
-- Table 0: FOM Failure Types (Taxonomy)
-- Enables targeted training on specific failure patterns
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_failure_types (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    failure_code        VARCHAR(50) NOT NULL UNIQUE,
    failure_class       VARCHAR(50) NOT NULL,
    recovery_action     VARCHAR(100),
    parent_failure_id   UUID REFERENCES fom_failure_types(id),
    description         TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- Seed failure taxonomy
INSERT INTO fom_failure_types (failure_code, failure_class, recovery_action, description) VALUES
    ('DOMAIN_MISSING', 'prerequisite', 'buy_domain_first', 'Domain not registered before deployment'),
    ('DNS_NOT_CONFIGURED', 'prerequisite', 'configure_dns', 'DNS records not set up'),
    ('AUTH_FAILED', 'auth', 'retry_with_credentials', 'Authentication credentials invalid or expired'),
    ('TIMEOUT', 'execution', 'retry_with_timeout', 'Function execution exceeded timeout limit'),
    ('RESOURCE_EXHAUSTED', 'resource', 'retry_with_more_resources', 'Out of memory or CPU'),
    ('DEPENDENCY_MISSING', 'prerequisite', 'install_dependency', 'Required dependency not installed'),
    ('PAYMENT_FAILED', 'resource', 'update_payment_method', 'Payment method declined'),
    ('RATE_LIMIT_EXCEEDED', 'resource', 'wait_and_retry', 'API rate limit hit'),
    ('INVALID_INPUT', 'execution', 'validate_input', 'Input validation failed'),
    ('NETWORK_ERROR', 'execution', 'retry', 'Network connectivity issue')
ON CONFLICT (failure_code) DO NOTHING;

-- ============================================================
-- Table 1: FOM Goals
-- Every user goal/request that triggers a workflow
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_goals (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID NOT NULL,
    user_id                 UUID NOT NULL,
    goal_text               TEXT NOT NULL,
    goal_type               VARCHAR(50) NOT NULL,
    goal_category           VARCHAR(50) NOT NULL,
    context                 JSONB,
    constraints             JSONB,
    user_tier               VARCHAR(20) DEFAULT 'free',
    user_experience_level   VARCHAR(20) DEFAULT 'intermediate',
    user_domain             VARCHAR(50),
    user_goals_history_count INT DEFAULT 0,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    source                  VARCHAR(20) DEFAULT 'user'
);

CREATE INDEX IF NOT EXISTS idx_fom_goals_tenant ON fom_goals(tenant_id);
CREATE INDEX IF NOT EXISTS idx_fom_goals_type ON fom_goals(goal_type);
CREATE INDEX IF NOT EXISTS idx_fom_goals_created ON fom_goals(created_at);
CREATE INDEX IF NOT EXISTS idx_fom_goals_user ON fom_goals(user_id);

-- ============================================================
-- Table 2: FOM Plans
-- Workflow plans generated to achieve a goal
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_plans (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id             UUID NOT NULL REFERENCES fom_goals(id),
    plan_text           TEXT NOT NULL,
    workflow_json       JSONB NOT NULL,
    model_used          VARCHAR(100),
    generation_time_ms  INT,
    confidence          DECIMAL(5,4),
    estimated_cost      DECIMAL(10,6),
    estimated_time      INT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fom_plans_goal ON fom_plans(goal_id);

-- ============================================================
-- Table 3: FOM Actions
-- Individual function calls within a plan execution
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_actions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id         UUID NOT NULL REFERENCES fom_plans(id),
    function_name   VARCHAR(255) NOT NULL,
    function_id     UUID,
    input_schema    JSONB,
    output_schema   JSONB,
    execution_id    UUID,
    actual_cost     DECIMAL(10,6),
    actual_time_ms  INT,
    success         BOOLEAN,
    error_message   TEXT,
    sequence_order  INT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fom_actions_plan ON fom_actions(plan_id);
CREATE INDEX IF NOT EXISTS idx_fom_actions_function ON fom_actions(function_name);

-- ============================================================
-- Table 4: FOM Results
-- Outcome of executing a plan
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_results (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id                 UUID NOT NULL REFERENCES fom_plans(id),
    success                 BOOLEAN NOT NULL,
    outcome_text            TEXT,
    total_cost              DECIMAL(10,6) NOT NULL,
    total_time_ms           INT NOT NULL,
    reliability_score       INT DEFAULT 0,
    efficiency_score        INT DEFAULT 0,
    speed_score             INT DEFAULT 0,
    completeness_score      INT DEFAULT 0,
    failure_reason          TEXT,
    failure_code            VARCHAR(50) REFERENCES fom_failure_types(failure_code),
    failed_action_id        UUID REFERENCES fom_actions(id),
    user_rating             INT,
    user_feedback           TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fom_results_plan ON fom_results(plan_id);
CREATE INDEX IF NOT EXISTS idx_fom_results_success ON fom_results(success);
CREATE INDEX IF NOT EXISTS idx_fom_results_created ON fom_results(created_at);

-- ============================================================
-- Table 5: FOM Workflow Patterns
-- Stores workflow patterns for learning
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_workflow_patterns (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern_name        VARCHAR(255) NOT NULL,
    goal_type           VARCHAR(50) NOT NULL,
    workflow_json       JSONB NOT NULL,
    usage_count         INT DEFAULT 0,
    success_count       INT DEFAULT 0,
    failure_count       INT DEFAULT 0,
    avg_cost            DECIMAL(10,6),
    avg_time_ms         INT,
    avg_success_rate    DECIMAL(5,4),
    first_used_at       TIMESTAMPTZ,
    last_used_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(pattern_name, goal_type)
);

CREATE INDEX IF NOT EXISTS idx_fom_patterns_goal_type ON fom_workflow_patterns(goal_type);

-- ============================================================
-- Table 6: FOM Function Stats
-- Per-function statistics for physics engine
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_function_stats (
    function_name   VARCHAR(255) PRIMARY KEY,
    avg_cost        DECIMAL(10,6) NOT NULL DEFAULT 0,
    avg_time_ms     INT NOT NULL DEFAULT 0,
    success_rate    DECIMAL(5,4) NOT NULL DEFAULT 0.95,
    p50_time_ms     INT,
    p95_time_ms     INT,
    p99_time_ms     INT,
    dependencies    TEXT[],
    sample_count    INT DEFAULT 0,
    last_updated    TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- Table 7: FOM Training Records
-- Denormalized for efficient training data export
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_training_records (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_text           TEXT NOT NULL,
    goal_type           VARCHAR(50) NOT NULL,
    workflow_json       JSONB NOT NULL,
    outcome_success     BOOLEAN NOT NULL,
    outcome_score       INT NOT NULL,
    total_cost          DECIMAL(10,6),
    total_time_ms       INT,
    is_synthetic        BOOLEAN DEFAULT FALSE,
    generation_method   VARCHAR(50),
    data_source         VARCHAR(50) DEFAULT 'production',
    confidence_level    VARCHAR(20) DEFAULT 'high',
    labeled_by          UUID,
    labeling_method     VARCHAR(50),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    split               VARCHAR(10) DEFAULT 'train'
);

CREATE INDEX IF NOT EXISTS idx_fom_training_goal_type ON fom_training_records(goal_type);
CREATE INDEX IF NOT EXISTS idx_fom_training_split ON fom_training_records(split);
CREATE INDEX IF NOT EXISTS idx_fom_training_created ON fom_training_records(created_at);

-- ============================================================
-- Table 8: FOM Events (Event Sourcing)
-- Full execution trace for replay, debugging, and sequence training
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id    UUID NOT NULL,
    plan_id         UUID REFERENCES fom_plans(id),
    event_type      VARCHAR(50) NOT NULL,
    timestamp       TIMESTAMPTZ DEFAULT NOW(),
    payload         JSONB NOT NULL,
    sequence_order  INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_fom_events_execution ON fom_events(execution_id);
CREATE INDEX IF NOT EXISTS idx_fom_events_plan ON fom_events(plan_id);
CREATE INDEX IF NOT EXISTS idx_fom_events_type ON fom_events(event_type);
CREATE INDEX IF NOT EXISTS idx_fom_events_timestamp ON fom_events(timestamp);

-- ============================================================
-- Table 9: FOM Workflow Hints
-- When to use which workflow (context-aware recommendations)
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_workflow_hints (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_pattern        TEXT NOT NULL,
    goal_type           VARCHAR(50) NOT NULL,
    context_conditions  JSONB,
    recommended_workflow JSONB NOT NULL,
    success_rate        DECIMAL(5,4),
    avg_time_savings_ms INT,
    constraint_tags     TEXT[],
    hint_usage_count    INT DEFAULT 0,
    hint_success_count  INT DEFAULT 0,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(goal_pattern, goal_type, context_conditions)
);

CREATE INDEX IF NOT EXISTS idx_fom_hints_goal_type ON fom_workflow_hints(goal_type);

-- ============================================================
-- Table 10: FOM Privacy Budget
-- Tracks privacy spend to prevent over-training on specific users
-- ============================================================
CREATE TABLE IF NOT EXISTS fom_privacy_budget (
    tenant_id       UUID PRIMARY KEY,
    total_budget    INT NOT NULL DEFAULT 10000,
    used_budget     INT NOT NULL DEFAULT 0,
    period_start    TIMESTAMPTZ NOT NULL,
    period_end      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- Down Migration
-- ============================================================
-- DROP TABLE IF EXISTS fom_privacy_budget;
-- DROP TABLE IF EXISTS fom_workflow_hints;
-- DROP TABLE IF EXISTS fom_events;
-- DROP TABLE IF EXISTS fom_training_records;
-- DROP TABLE IF EXISTS fom_function_stats;
-- DROP TABLE IF EXISTS fom_workflow_patterns;
-- DROP TABLE IF EXISTS fom_results;
-- DROP TABLE IF EXISTS fom_actions;
-- DROP TABLE IF EXISTS fom_plans;
-- DROP TABLE IF EXISTS fom_goals;
-- DROP TABLE IF EXISTS fom_failure_types;