CREATE TABLE IF NOT EXISTS marketplace_ratings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    extension_id UUID NOT NULL REFERENCES marketplace_extensions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    rating SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(extension_id, tenant_id)
);

CREATE INDEX idx_marketplace_ratings_extension ON marketplace_ratings(extension_id);
CREATE INDEX idx_marketplace_ratings_tenant ON marketplace_ratings(tenant_id);
CREATE INDEX idx_marketplace_ratings_created ON marketplace_ratings(created_at DESC);

COMMENT ON TABLE marketplace_ratings IS 'Per-tenant ratings and reviews for marketplace extensions';
