-- Marketplace Extensions Table
-- Extension marketplace listings without monetization

CREATE TABLE IF NOT EXISTS marketplace_extensions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL,
    plugin_id UUID REFERENCES plugins(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    icon_url VARCHAR(500),
    screenshots TEXT[],
    manifest JSONB NOT NULL,
    manifest_url VARCHAR(500),
    signature TEXT,
    verified BOOLEAN DEFAULT false,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    featured BOOLEAN DEFAULT false,
    install_count INTEGER DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0,
    rating_count INTEGER DEFAULT 0,
    trust_score DECIMAL(5,2) DEFAULT 0,
    sandbox_score DECIMAL(5,2) DEFAULT 0,
    security_score DECIMAL(5,2) DEFAULT 0,
    runtime_score DECIMAL(5,2) DEFAULT 0,
    compatibility JSONB DEFAULT '{}',
    tags TEXT[],
    changelog TEXT,
    release_notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    unpublished_at TIMESTAMPTZ,
    UNIQUE(creator_id, name)
);

CREATE INDEX idx_marketplace_extensions_creator ON marketplace_extensions(creator_id);
CREATE INDEX idx_marketplace_extensions_category ON marketplace_extensions(category);
CREATE INDEX idx_marketplace_extensions_status ON marketplace_extensions(status);
CREATE INDEX idx_marketplace_extensions_featured ON marketplace_extensions(featured);
CREATE INDEX idx_marketplace_extensions_install_count ON marketplace_extensions(install_count DESC);
CREATE INDEX idx_marketplace_extensions_rating ON marketplace_extensions(rating_average DESC);
CREATE INDEX idx_marketplace_extensions_created ON marketplace_extensions(created_at DESC);

COMMENT ON TABLE marketplace_extensions IS 'Marketplace extension listings - discovery and distribution without monetization';