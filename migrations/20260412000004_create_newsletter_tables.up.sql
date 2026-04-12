-- Create newsletter subscribers table
CREATE TABLE IF NOT EXISTS newsletter_subscribers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'unsubscribed', 'bounced')),
    source VARCHAR(50),
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    subscribed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    unsubscribed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for active subscribers lookup
CREATE INDEX idx_newsletter_subscribers_status ON newsletter_subscribers(status);
CREATE INDEX idx_newsletter_subscribers_email ON newsletter_subscribers(email);
CREATE INDEX idx_newsletter_subscribers_subscribed_at ON newsletter_subscribers(subscribed_at);

-- Create newsletter campaigns table
CREATE TABLE IF NOT EXISTS newsletter_campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject VARCHAR(255) NOT NULL,
    preview_text VARCHAR(255),
    content TEXT NOT NULL,
    html_content TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'sent', 'failed')),
    sent_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    sent_count INTEGER NOT NULL DEFAULT 0,
    open_count INTEGER NOT NULL DEFAULT 0,
    click_count INTEGER NOT NULL DEFAULT 0,
    bounce_count INTEGER NOT NULL DEFAULT 0,
    error_message VARCHAR(500),
    metadata JSONB,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for campaigns
CREATE INDEX idx_newsletter_campaigns_status ON newsletter_campaigns(status);
CREATE INDEX idx_newsletter_campaigns_created_by ON newsletter_campaigns(created_by);
CREATE INDEX idx_newsletter_campaigns_created_at ON newsletter_campaigns(created_at);

-- Create newsletter campaign emails tracking table
CREATE TABLE IF NOT EXISTS newsletter_campaign_emails (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES newsletter_campaigns(id) ON DELETE CASCADE,
    subscriber_id UUID NOT NULL REFERENCES newsletter_subscribers(id) ON DELETE CASCADE,
    email_id VARCHAR(255), -- Resend email ID for tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'delivered', 'opened', 'clicked', 'bounced', 'complained')),
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    opened_at TIMESTAMP WITH TIME ZONE,
    clicked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for campaign emails
CREATE INDEX idx_newsletter_campaign_emails_campaign_id ON newsletter_campaign_emails(campaign_id);
CREATE INDEX idx_newsletter_campaign_emails_subscriber_id ON newsletter_campaign_emails(subscriber_id);
CREATE INDEX idx_newsletter_campaign_emails_status ON newsletter_campaign_emails(status);

-- Create unique index to prevent duplicate emails per campaign
CREATE UNIQUE INDEX idx_newsletter_campaign_emails_unique ON newsletter_campaign_emails(campaign_id, subscriber_id);

-- Enable row level security (consistent with other tables)
ALTER TABLE newsletter_subscribers ENABLE ROW LEVEL SECURITY;
ALTER TABLE newsletter_campaigns ENABLE ROW LEVEL SECURITY;
ALTER TABLE newsletter_campaign_emails ENABLE ROW LEVEL SECURITY;

-- Create policies for admin access
CREATE POLICY newsletter_subscribers_admin_policy ON newsletter_subscribers
    FOR ALL TO admin_role
    USING (true);

CREATE POLICY newsletter_campaigns_admin_policy ON newsletter_campaigns
    FOR ALL TO admin_role
    USING (true);

CREATE POLICY newsletter_campaign_emails_admin_policy ON newsletter_campaign_emails
    FOR ALL TO admin_role
    USING (true);

-- Create trigger for updating updated_at timestamp
CREATE OR REPLACE FUNCTION update_newsletter_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_newsletter_subscribers_updated_at
    BEFORE UPDATE ON newsletter_subscribers
    FOR EACH ROW
    EXECUTE FUNCTION update_newsletter_updated_at();

CREATE TRIGGER trigger_newsletter_campaigns_updated_at
    BEFORE UPDATE ON newsletter_campaigns
    FOR EACH ROW
    EXECUTE FUNCTION update_newsletter_updated_at();

CREATE TRIGGER trigger_newsletter_campaign_emails_updated_at
    BEFORE UPDATE ON newsletter_campaign_emails
    FOR EACH ROW
    EXECUTE FUNCTION update_newsletter_updated_at();
