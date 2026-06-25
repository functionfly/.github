-- Founder Tiers table for graduated benefits
CREATE TABLE IF NOT EXISTS founder_tiers (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    tier TEXT NOT NULL DEFAULT 'founder',
    referral_count INT DEFAULT 0,
    total_earnings_cents BIGINT DEFAULT 0,
    rank INT,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for tier queries
CREATE INDEX IF NOT EXISTS idx_founder_tiers_tier ON founder_tiers(tier);
CREATE INDEX IF NOT EXISTS idx_founder_tiers_rank ON founder_tiers(rank);
CREATE INDEX IF NOT EXISTS idx_founder_tiers_earnings ON founder_tiers(total_earnings_cents DESC);
