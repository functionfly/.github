-- Browser Automation Migration
-- Creates tables for browser automation capability

-- Agent browser configuration
CREATE TABLE IF NOT EXISTS agent_browser_configs (
    agent_id VARCHAR(255) PRIMARY KEY,
    browser_enabled BOOLEAN NOT NULL DEFAULT true,
    allowed_domains TEXT[],
    max_sessions INTEGER NOT NULL DEFAULT 1,
    credential_storage_enabled BOOLEAN NOT NULL DEFAULT false,
    default_timeout_ms INTEGER NOT NULL DEFAULT 30000,
    headful_mode BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Agent browser sessions
CREATE TABLE IF NOT EXISTS agent_browser_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL,
    session_type VARCHAR(50) NOT NULL DEFAULT 'shared',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    browser_port INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    closed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_agent_browser_sessions_agent_id ON agent_browser_sessions(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_browser_sessions_status ON agent_browser_sessions(status);

-- Agent browser usage tracking for cost attribution
CREATE TABLE IF NOT EXISTS agent_browser_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL,
    session_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    domain VARCHAR(255),
    duration_ms INTEGER,
    browser_minutes DECIMAL(10,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_browser_usage_agent_id ON agent_browser_usage(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_browser_usage_created_at ON agent_browser_usage(created_at);

-- Agent browser credentials (encrypted)
CREATE TABLE IF NOT EXISTS agent_browser_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    encrypted_data BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_browser_credentials_agent_id ON agent_browser_credentials(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_browser_credentials_domain ON agent_browser_credentials(domain);

-- Agent browser dead letters (failed sessions for inspection)
CREATE TABLE IF NOT EXISTS agent_browser_dead_letters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) NOT NULL,
    agent_id VARCHAR(255) NOT NULL,
    error_type VARCHAR(50) NOT NULL,
    error_message TEXT,
    last_url VARCHAR(2048),
    cookies TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_browser_dead_letters_agent_id ON agent_browser_dead_letters(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_browser_dead_letters_created_at ON agent_browser_dead_letters(created_at);
