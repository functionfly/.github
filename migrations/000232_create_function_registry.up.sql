-- FunctionFly Public Function Registry Schema
-- This creates the foundation for globally addressable functions

-- 1. Functions table - Identity and metadata layer
CREATE TABLE IF NOT EXISTS registry_functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    latest_version VARCHAR(20),
    title VARCHAR(255),
    description TEXT,
    category VARCHAR(50),
    tags JSONB DEFAULT '[]'::jsonb,
    visibility VARCHAR(20) DEFAULT 'public' CHECK (visibility IN ('public', 'private', 'unlisted')),

    -- Pricing (MVP)
    price_per_call NUMERIC(20, 8) DEFAULT 0,

    -- Trust & Discovery scores
    popularity_score INTEGER DEFAULT 0,
    reliability_score NUMERIC(5, 2) DEFAULT 0,
    deterministic_score NUMERIC(5, 2) DEFAULT 0,

    -- Ownership
    tenant_id UUID REFERENCES tenants(id),
    owner_user_id UUID REFERENCES users(id),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_registry_author_name UNIQUE (author, name)
);

-- Indexes for functions
CREATE INDEX IF NOT EXISTS idx_registry_functions_author ON registry_functions(author);
CREATE INDEX IF NOT EXISTS idx_registry_functions_category ON registry_functions(category);
CREATE INDEX IF NOT EXISTS idx_registry_functions_visibility ON registry_functions(visibility);
CREATE INDEX IF NOT EXISTS idx_registry_functions_tags ON registry_functions USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_registry_functions_name_search ON registry_functions USING gin(to_tsvector('english', name || ' ' || COALESCE(title, '') || ' ' || COALESCE(description, '')));


-- 2. Function versions - Immutable version storage
CREATE TABLE IF NOT EXISTS registry_function_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    version VARCHAR(20) NOT NULL,

    -- Manifest data (the functionfly.json contents)
    manifest JSONB NOT NULL,

    -- Execution characteristics
    runtime VARCHAR(50) NOT NULL,
    timeout_ms INTEGER DEFAULT 5000,
    memory_mb INTEGER DEFAULT 128,
    deterministic BOOLEAN DEFAULT false,
    cache_ttl INTEGER DEFAULT 0,

    -- Deployment target for this version
    deployment_id UUID REFERENCES deployments(id),
    backend_id UUID REFERENCES backends(id),

    -- Content hash for integrity verification
    content_hash VARCHAR(64),

    published_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_registry_function_version UNIQUE (function_id, version)
);

-- Indexes for versions
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_function_id ON registry_function_versions(function_id);
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_published_at ON registry_function_versions(published_at);
CREATE INDEX IF NOT EXISTS idx_registry_function_versions_runtime ON registry_function_versions(runtime);


-- 3. Function executions - Tracking for stats & billing
CREATE TABLE IF NOT EXISTS registry_function_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id),
    version VARCHAR(20) NOT NULL,

    -- Execution details
    duration_ms INTEGER NOT NULL,
    status_code INTEGER NOT NULL,
    cached BOOLEAN DEFAULT false,

    -- Outcome
    outcome VARCHAR(20) NOT NULL CHECK (outcome IN ('success', 'error', 'timeout')),
    error_code VARCHAR(50),

    -- Request metadata (for rate limiting and analytics)
    caller_ip INET,
    user_agent TEXT,
    geo_country VARCHAR(2),

    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for executions
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_function_id ON registry_function_executions(function_id);
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_timestamp ON registry_function_executions(timestamp);
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_outcome ON registry_function_executions(outcome);
CREATE INDEX IF NOT EXISTS idx_registry_function_executions_caller_ip ON registry_function_executions(caller_ip);


-- 4. Function ratings - Trust layer
CREATE TABLE IF NOT EXISTS registry_function_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Rating scores (0-5 stars stored as 0-100)
    overall_score NUMERIC(5, 2) DEFAULT 0,
    reliability_score NUMERIC(5, 2) DEFAULT 0,
    latency_score NUMERIC(5, 2) DEFAULT 0,
    documentation_score NUMERIC(5, 2) DEFAULT 0,

    -- Aggregated stats
    total_ratings INTEGER DEFAULT 0,
    success_rate NUMERIC(5, 2) DEFAULT 0,
    p95_latency_ms INTEGER DEFAULT 0,
    avg_latency_ms INTEGER DEFAULT 0,

    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_registry_function_rating UNIQUE (function_id)
);

CREATE INDEX IF NOT EXISTS idx_registry_function_ratings_function_id ON registry_function_ratings(function_id);
CREATE INDEX IF NOT EXISTS idx_registry_function_ratings_score ON registry_function_ratings(overall_score DESC);
