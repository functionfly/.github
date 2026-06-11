-- AEP Phase 4: Agent Swarm Architecture
-- Migration: 000058_aep_phase4_swarm_architecture
-- This adds hierarchical swarm coordination, messaging, and wallet support

-- ============================================
-- 1. Extend agent_identities with swarm fields
-- ============================================

ALTER TABLE agent_identities 
ADD COLUMN IF NOT EXISTS parent_agent_id TEXT,
ADD COLUMN IF NOT EXISTS swarm_role TEXT DEFAULT 'worker' CHECK (swarm_role IN ('worker', 'manager', 'infrastructure')),
ADD COLUMN IF NOT EXISTS max_child_agents INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS capabilities JSONB DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS autonomous_enabled BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS evolution_enabled BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS trust_score NUMERIC(5,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS economic_score NUMERIC(5,2) DEFAULT 0;

-- Index for parent-child lookups
CREATE INDEX IF NOT EXISTS idx_agent_identities_parent_id ON agent_identities(parent_agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_identities_swarm_role ON agent_identities(swarm_role);

-- ============================================
-- 2. Agent relationships table (DAG structure)
-- ============================================

CREATE TABLE IF NOT EXISTS agent_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_agent_id TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    child_agent_id TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL DEFAULT 'parent' CHECK (relationship_type IN ('parent', 'supervisor', 'collaborator')),
    max_delegation_depth INT NOT NULL DEFAULT 5,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_agent_relationship UNIQUE (parent_agent_id, child_agent_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_relationships_parent ON agent_relationships(parent_agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_relationships_child ON agent_relationships(child_agent_id);

-- ============================================
-- 3. Agent messages table (A2A communication)
-- ============================================

CREATE TABLE IF NOT EXISTS agent_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_agent_id TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    to_agent_id TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    message_type TEXT NOT NULL CHECK (message_type IN (
        'task_delegation', 'task_result', 'query', 'response', 
        'capability_discovery', 'heartbeat', 'evolution_proposal', 'budget_request'
    )),
    payload JSONB DEFAULT '{}'::jsonb,
    session_id TEXT,
    ttl_seconds INT NOT NULL DEFAULT 3600,  -- 1 hour default TTL
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'delivered', 'read', 'expired', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_agent_messages_to_agent ON agent_messages(to_agent_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_messages_from_agent ON agent_messages(from_agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_messages_session ON agent_messages(session_id);

-- ============================================
-- 4. Agent wallets table (Economic layer)
-- ============================================

CREATE TABLE IF NOT EXISTS agent_wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL UNIQUE REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    balance_usd DECIMAL(12,2) NOT NULL DEFAULT 0,
    escrow_balance_usd DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_earned_usd DECIMAL(12,2) NOT NULL DEFAULT 0,
    total_spent_usd DECIMAL(12,2) NOT NULL DEFAULT 0,
    last_earning_at TIMESTAMPTZ,
    last_spending_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_wallets_agent_id ON agent_wallets(agent_id);

-- ============================================
-- 5. Agent listings table (Marketplace - Agents)
-- ============================================

CREATE TABLE IF NOT EXISTS agent_listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL UNIQUE REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    listing_type TEXT NOT NULL DEFAULT 'worker' CHECK (listing_type IN ('worker', 'manager', 'infrastructure')),
    pricing_model TEXT NOT NULL DEFAULT 'per_call' CHECK (pricing_model IN ('free', 'per_call', 'subscription', 'revenue_share')),
    price_per_call DECIMAL(10,4),
    subscription_monthly_usd DECIMAL(10,2),
    revenue_share_percent DECIMAL(5,2),
    rating_score NUMERIC(3,2) DEFAULT 0,
    total_calls INT NOT NULL DEFAULT 0,
    roi_score NUMERIC(5,2) DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    listed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_listings_type ON agent_listings(listing_type);
CREATE INDEX IF NOT EXISTS idx_agent_listings_active ON agent_listings(is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_agent_listings_rating ON agent_listings(rating_score DESC);

-- ============================================
-- 6. Function listings table (Marketplace - Functions)
-- ============================================

CREATE TABLE IF NOT EXISTS function_listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    pricing_model TEXT NOT NULL DEFAULT 'per_call' CHECK (pricing_model IN ('free', 'per_call', 'subscription', 'revenue_share')),
    price_per_call DECIMAL(10,4),
    subscription_monthly_usd DECIMAL(10,2),
    revenue_share_percent DECIMAL(5,2),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    rating_score NUMERIC(3,2) DEFAULT 0,
    call_volume INT NOT NULL DEFAULT 0,
    deterministic_verified BOOLEAN NOT NULL DEFAULT FALSE,
    listed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_function_listings_function_id ON function_listings(function_id);
CREATE INDEX IF NOT EXISTS idx_function_listings_active ON function_listings(is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_function_listings_deterministic ON function_listings(deterministic_verified) WHERE deterministic_verified = TRUE;

-- ============================================
-- 7. Revenue transactions table (Economic enforcement)
-- ============================================

CREATE TABLE IF NOT EXISTS agent_revenue_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_agent_id TEXT REFERENCES agent_identities(agent_id) ON DELETE SET NULL,
    to_agent_id TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    function_id UUID REFERENCES registry_functions(id) ON DELETE SET NULL,
    amount_usd DECIMAL(10,4) NOT NULL,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('delegation_payment', 'function_call', 'subscription', 'revenue_share', 'refund')),
    session_id TEXT,
    execution_id TEXT,
    parent_execution_id TEXT,
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_revenue_to_agent ON agent_revenue_transactions(to_agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_revenue_session ON agent_revenue_transactions(session_id);
CREATE INDEX IF NOT EXISTS idx_agent_revenue_execution ON agent_revenue_transactions(execution_id);

-- ============================================
-- 8. Agent evolution proposals table
-- ============================================

CREATE TABLE IF NOT EXISTS agent_evolution_proposals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    proposal_type TEXT NOT NULL CHECK (proposal_type IN (
        'spawn_specialist', 'modify_policy', 'adjust_timeout', 
        'generate_function', 'retire_child', 'upgrade_capabilities'
    )),
    proposal_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'implemented', 'expired')),
    parent_approval_required BOOLEAN NOT NULL DEFAULT TRUE,
    simulated_outcome JSONB,
    approved_by TEXT,
    implemented_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_evolution_agent ON agent_evolution_proposals(agent_id, status);
CREATE INDEX IF NOT EXISTS idx_agent_evolution_type ON agent_evolution_proposals(proposal_type);

-- ============================================
-- 9. Agent autonomy schedules table
-- ============================================

CREATE TABLE IF NOT EXISTS agent_autonomy_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    schedule_type TEXT NOT NULL CHECK (schedule_type IN ('recurring', 'one_time', 'trigger_based')),
    cron_expression TEXT,
    trigger_event TEXT,
    trigger_condition JSONB,
    action_type TEXT NOT NULL CHECK (action_type IN ('execute_function', 'spawn_agent', 'send_message', 'update_state', 'evolve')),
    action_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    next_run_at TIMESTAMPTZ,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_autonomy_agent ON agent_autonomy_schedules(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_autonomy_active ON agent_autonomy_schedules(is_active, next_run_at) WHERE is_active = TRUE;

-- ============================================
-- 10. Function generation metadata
-- ============================================

ALTER TABLE registry_functions 
ADD COLUMN IF NOT EXISTS owner_agent_id TEXT REFERENCES agent_identities(agent_id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS agent_generated BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS generation_prompt_hash VARCHAR(64),
ADD COLUMN IF NOT EXISTS generation_model VARCHAR(100),
ADD COLUMN IF NOT EXISTS deterministic_cert_hash VARCHAR(64),
ADD COLUMN IF NOT EXISTS revenue_total_usd DECIMAL(12,2) DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_registry_functions_agent_owner ON registry_functions(owner_agent_id);
CREATE INDEX IF NOT EXISTS idx_registry_functions_agent_generated ON registry_functions(agent_generated);
