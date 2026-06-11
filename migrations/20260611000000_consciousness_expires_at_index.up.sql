-- Add index on expires_at for ExpireOldInsights query performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_consciousness_insights_expires_at
    ON consciousness_insights(expires_at) WHERE expires_at IS NOT NULL;