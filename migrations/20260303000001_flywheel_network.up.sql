-- Flywheel Network Database Schema
-- This migration creates all tables for the Flywheel Network feature

-- Flywheel Problems
CREATE TABLE flywheel_problems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,

    category VARCHAR(50) NOT NULL,
    tags JSONB DEFAULT '[]'::jsonb,
    difficulty VARCHAR(20) NOT NULL,

    environment_spec JSONB NOT NULL,
    test_cases JSONB NOT NULL,
    hidden_tests JSONB DEFAULT '[]'::jsonb,
    attachments JSONB DEFAULT '[]'::jsonb,

    capsule_context JSONB,
    ai_formatted BOOLEAN DEFAULT false,
    formatted_by UUID,

    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    visibility VARCHAR(20) NOT NULL DEFAULT 'public',
    bounty_amount NUMERIC(20, 8) DEFAULT 0,

    view_count BIGINT DEFAULT 0,
    solution_count INT DEFAULT 0,
    success_rate NUMERIC(5, 2) DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_flywheel_problems_slug UNIQUE (slug)
);

CREATE INDEX idx_flywheel_problems_author ON flywheel_problems(author_id);
CREATE INDEX idx_flywheel_problems_category ON flywheel_problems(category);
CREATE INDEX idx_flywheel_problems_difficulty ON flywheel_problems(difficulty);
CREATE INDEX idx_flywheel_problems_status ON flywheel_problems(status);
CREATE INDEX idx_flywheel_problems_tags ON flywheel_problems USING GIN(tags);
CREATE INDEX idx_flywheel_problems_search ON flywheel_problems
    USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- Flywheel Solutions
CREATE TABLE flywheel_solutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    problem_id UUID NOT NULL REFERENCES flywheel_problems(id),
    author_id UUID NOT NULL REFERENCES users(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    parent_id UUID REFERENCES flywheel_solutions(id),

    type VARCHAR(20) NOT NULL, -- code, capsule, agent_fork, patch

    code_solution JSONB,
    capsule_solution JSONB,
    agent_fork JSONB,
    patch_solution JSONB,

    verification_result JSONB,
    benchmark_results JSONB,
    compute_cost JSONB,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    visibility VARCHAR(20) NOT NULL DEFAULT 'public',

    reputation_delta JSONB,
    marketplace_uri VARCHAR(255),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    submitted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_flywheel_solutions_problem ON flywheel_solutions(problem_id);
CREATE INDEX idx_flywheel_solutions_author ON flywheel_solutions(author_id);
CREATE INDEX idx_flywheel_solutions_parent ON flywheel_solutions(parent_id);
CREATE INDEX idx_flywheel_solutions_type ON flywheel_solutions(type);
CREATE INDEX idx_flywheel_solutions_status ON flywheel_solutions(status);

-- Reputation Profiles
CREATE TABLE reputation_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    builder_score INT DEFAULT 0,
    optimizer_score INT DEFAULT 0,
    mentor_score INT DEFAULT 0,
    agent_whisperer_score INT DEFAULT 0,

    reliability_index NUMERIC(5, 4) DEFAULT 1.0,
    consistency_score NUMERIC(5, 4) DEFAULT 1.0,
    overall_score INT DEFAULT 0,

    tier VARCHAR(20) DEFAULT 'novice',
    badges JSONB DEFAULT '[]'::jsonb,

    stats JSONB DEFAULT '{}'::jsonb,
    score_history JSONB DEFAULT '[]'::jsonb,

    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_reputation_overall_score ON reputation_profiles(overall_score DESC);
CREATE INDEX idx_reputation_tier ON reputation_profiles(tier);

-- Agent Attachments
CREATE TABLE agent_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL,
    agent_id VARCHAR(255) NOT NULL,
    agent_name VARCHAR(255) NOT NULL,
    agent_owner_id UUID NOT NULL REFERENCES users(id),

    role VARCHAR(20) NOT NULL,
    capabilities JSONB DEFAULT '[]'::jsonb,
    context_snapshot JSONB,
    system_prompt TEXT,

    status VARCHAR(20) DEFAULT 'active',
    messages_sent INT DEFAULT 0,
    solutions_proposed INT DEFAULT 0,

    attached_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_agent_attachments_thread ON agent_attachments(thread_id);
CREATE INDEX idx_agent_attachments_agent ON agent_attachments(agent_id);

-- Debates
CREATE TABLE debates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL,
    problem_id UUID REFERENCES flywheel_problems(id),

    topic VARCHAR(255) NOT NULL,
    format VARCHAR(20) NOT NULL,

    current_round INT DEFAULT 0,
    total_rounds INT NOT NULL,
    rounds JSONB DEFAULT '[]'::jsonb,

    status VARCHAR(20) DEFAULT 'pending',
    winner JSONB,
    consensus JSONB,

    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_debates_thread ON debates(thread_id);
CREATE INDEX idx_debates_status ON debates(status);

-- Challenges
CREATE TABLE challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    problem_id UUID NOT NULL REFERENCES flywheel_problems(id),

    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,

    type VARCHAR(20) NOT NULL,
    scoring_config JSONB NOT NULL,
    rewards JSONB NOT NULL,

    max_participants INT,
    min_participants INT DEFAULT 1,

    status VARCHAR(20) DEFAULT 'upcoming',
    participant_count INT DEFAULT 0,
    submission_count INT DEFAULT 0,

    winners JSONB,
    leaderboard JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_challenges_status ON challenges(status);
CREATE INDEX idx_challenges_time ON challenges(start_time, end_time);

-- Executable Threads
CREATE TABLE flywheel_threads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES flywheel_threads(id),

    problem_id UUID REFERENCES flywheel_problems(id),
    challenge_id UUID REFERENCES challenges(id),

    title VARCHAR(255) NOT NULL,
    description TEXT,

    creator_id UUID NOT NULL REFERENCES users(id),
    participants JSONB DEFAULT '[]'::jsonb,
    agents JSONB DEFAULT '[]'::jsonb,

    message_count INT DEFAULT 0,
    executions JSONB DEFAULT '[]'::jsonb,

    version_info JSONB,
    versions JSONB DEFAULT '[]'::jsonb,

    status VARCHAR(20) DEFAULT 'active',
    visibility VARCHAR(20) DEFAULT 'public',

    fork_count INT DEFAULT 0,
    forked_from UUID,

    view_count BIGINT DEFAULT 0,
    star_count INT DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_flywheel_threads_creator ON flywheel_threads(creator_id);
CREATE INDEX idx_flywheel_threads_problem ON flywheel_threads(problem_id);
CREATE INDEX idx_flywheel_threads_challenge ON flywheel_threads(challenge_id);
CREATE INDEX idx_flywheel_threads_status ON flywheel_threads(status);

-- Thread Messages
CREATE TABLE flywheel_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES flywheel_threads(id),

    author_type VARCHAR(20) NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    author_name VARCHAR(255) NOT NULL,

    type VARCHAR(20) NOT NULL,
    content TEXT,
    metadata JSONB,

    language VARCHAR(50),
    code TEXT,
    solution_id UUID,
    execution_id UUID,

    reply_to UUID,
    mentions JSONB DEFAULT '[]'::jsonb,
    reactions JSONB DEFAULT '[]'::jsonb,

    dre_proof JSONB,
    edit_history JSONB DEFAULT '[]'::jsonb,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    edited_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_flywheel_messages_thread ON flywheel_messages(thread_id);
CREATE INDEX idx_flywheel_messages_author ON flywheel_messages(author_id);
CREATE INDEX idx_flywheel_messages_type ON flywheel_messages(type);
CREATE INDEX idx_flywheel_messages_created ON flywheel_messages(created_at);

-- Challenge Submissions
CREATE TABLE challenge_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID NOT NULL REFERENCES challenges(id),
    user_id UUID NOT NULL REFERENCES users(id),
    solution_id UUID NOT NULL REFERENCES flywheel_solutions(id),

    status VARCHAR(20) DEFAULT 'pending',

    primary_score NUMERIC(10, 4),
    secondary_scores JSONB,
    composite_score NUMERIC(10, 4),

    test_results JSONB,
    benchmarks JSONB,

    current_rank INT,
    previous_rank INT,

    disqualified BOOLEAN DEFAULT false,
    disqualify_reason TEXT,

    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    evaluated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_challenge_submissions_challenge ON challenge_submissions(challenge_id);
CREATE INDEX idx_challenge_submissions_user ON challenge_submissions(user_id);
CREATE INDEX idx_challenge_submissions_score ON challenge_submissions(composite_score DESC);

-- Agent Reputation
CREATE TABLE agent_reputations (
    agent_id VARCHAR(255) PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id),

    tasks_completed BIGINT DEFAULT 0,
    success_rate NUMERIC(5, 4) DEFAULT 0,

    debates_won INT DEFAULT 0,
    debates_participated INT DEFAULT 0,

    forks_created INT DEFAULT 0,
    successful_forks INT DEFAULT 0,
    fork_adoption_rate NUMERIC(5, 4) DEFAULT 0,

    solutions_proposed INT DEFAULT 0,
    solutions_accepted INT DEFAULT 0,
    avg_solution_score NUMERIC(5, 2) DEFAULT 0,

    avg_rounds_to_solution NUMERIC(5, 2) DEFAULT 0,
    avg_compute_used NUMERIC(20, 8) DEFAULT 0,

    trust_score NUMERIC(5, 2) DEFAULT 0,
    specializations JSONB DEFAULT '[]'::jsonb,

    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_agent_reputations_owner ON agent_reputations(owner_id);
CREATE INDEX idx_agent_reputations_trust ON agent_reputations(trust_score DESC);

-- Replays
CREATE TABLE flywheel_replays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_thread_id UUID NOT NULL REFERENCES flywheel_threads(id),

    replay_type VARCHAR(20) NOT NULL,
    messages JSONB NOT NULL,
    executions JSONB DEFAULT '[]'::jsonb,
    divergences JSONB DEFAULT '[]'::jsonb,

    new_thread_id UUID REFERENCES flywheel_threads(id),

    replay_duration_ms BIGINT,
    executions_run INT DEFAULT 0,
    cached_results_used INT DEFAULT 0,

    replayed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    replayed_by UUID NOT NULL REFERENCES users(id)
);

CREATE INDEX idx_flywheel_replays_original ON flywheel_replays(original_thread_id);
CREATE INDEX idx_flywheel_replays_new_thread ON flywheel_replays(new_thread_id);

-- Abuse Tracking
CREATE TABLE abuse_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),

    analysis_result JSONB NOT NULL,
    risk_score NUMERIC(5, 4) NOT NULL,
    risk_level VARCHAR(20) NOT NULL,

    enforcement_action VARCHAR(50),
    action_taken_at TIMESTAMP WITH TIME ZONE,

    reviewed_by UUID REFERENCES users(id),
    review_notes TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_abuse_tracking_user ON abuse_tracking(user_id);
CREATE INDEX idx_abuse_tracking_risk ON abuse_tracking(risk_level);

-- Suspensions
CREATE TABLE suspensions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),

    reason TEXT NOT NULL,
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ends_at TIMESTAMP WITH TIME ZONE NOT NULL,

    status VARCHAR(20) DEFAULT 'active',
    lifted_by UUID REFERENCES users(id),
    lifted_at TIMESTAMP WITH TIME ZONE,
    lift_reason TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_suspensions_user ON suspensions(user_id);
CREATE INDEX idx_suspensions_status ON suspensions(status);
CREATE INDEX idx_suspensions_time ON suspensions(starts_at, ends_at);
