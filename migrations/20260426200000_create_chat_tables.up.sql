-- Chat tables for /chat feature
-- Stores chat sessions and messages with AI

CREATE TABLE IF NOT EXISTS chat_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    title VARCHAR(255),
    model VARCHAR(100) DEFAULT 'gpt-4o-mini',
    connector_ids JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_sessions_tenant_user ON chat_sessions(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated_at ON chat_sessions(updated_at DESC);

CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system', 'function')),
    content TEXT,
    attachments JSONB DEFAULT '[]',
    model VARCHAR(100),
    tokens_used INT DEFAULT 0,
    latency_ms INT DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id, created_at);

CREATE TABLE IF NOT EXISTS chat_connectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    icon VARCHAR(50) DEFAULT 'plug',
    config JSONB NOT NULL DEFAULT '{}',
    encrypted_credentials TEXT,
    is_active BOOLEAN DEFAULT true,
    last_tested_at TIMESTAMPTZ,
    last_test_success BOOLEAN,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_connectors_tenant_user ON chat_connectors(tenant_id, user_id);

CREATE TABLE IF NOT EXISTS chat_functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    session_id UUID REFERENCES chat_sessions(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    code TEXT NOT NULL,
    language VARCHAR(50) DEFAULT 'typescript',
    connector_id UUID REFERENCES chat_connectors(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_functions_tenant_user ON chat_functions(tenant_id, user_id);

-- RAG: store embedded connector data for semantic search
CREATE TABLE IF NOT EXISTS chat_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    connector_id UUID REFERENCES chat_connectors(id) ON DELETE CASCADE,
    session_id UUID REFERENCES chat_sessions(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    embedding VECTOR(1536),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_embeddings_tenant ON chat_embeddings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_chat_embeddings_connector ON chat_embeddings(connector_id);

-- Billing: track token usage for chat
CREATE TABLE IF NOT EXISTS chat_billing_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    session_id UUID REFERENCES chat_sessions(id) ON DELETE SET NULL,
    tokens_used INT NOT NULL,
    model VARCHAR(100) NOT NULL,
    charged_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(session_id, charged_at)
);

CREATE INDEX IF NOT EXISTS idx_chat_billing_tenant_period ON chat_billing_adjustments(tenant_id, charged_at);