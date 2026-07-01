CREATE TABLE IF NOT EXISTS recommendation_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    function_id UUID NOT NULL,
    interaction_type VARCHAR(32) NOT NULL,
    context JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rec_interactions_user
    ON recommendation_interactions(tenant_id, user_id, created_at);

CREATE INDEX IF NOT EXISTS idx_rec_interactions_function
    ON recommendation_interactions(function_id, created_at);

CREATE INDEX IF NOT EXISTS idx_rec_interactions_type
    ON recommendation_interactions(interaction_type, created_at);
