-- Payout settings for founders
CREATE TABLE IF NOT EXISTS founder_payout_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    stripe_account_id TEXT,
    payout_threshold_cents INT DEFAULT 2500,
    payout_schedule TEXT DEFAULT 'monthly',
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_founder_payout_settings_user ON founder_payout_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_founder_payout_settings_active ON founder_payout_settings(is_active) WHERE is_active = true;

-- Newsletter subscriptions
CREATE TABLE IF NOT EXISTS founder_newsletter_subscriptions (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    is_subscribed BOOLEAN DEFAULT true,
    subscribed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    unsubscribed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_founder_newsletter_subscriptions_subscribed ON founder_newsletter_subscriptions(is_subscribed) WHERE is_subscribed = true;
