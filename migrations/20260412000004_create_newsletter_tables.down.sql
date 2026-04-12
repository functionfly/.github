-- Drop triggers
DROP TRIGGER IF EXISTS trigger_newsletter_campaign_emails_updated_at ON newsletter_campaign_emails;
DROP TRIGGER IF EXISTS trigger_newsletter_campaigns_updated_at ON newsletter_campaigns;
DROP TRIGGER IF EXISTS trigger_newsletter_subscribers_updated_at ON newsletter_subscribers;

-- Drop function
DROP FUNCTION IF EXISTS update_newsletter_updated_at();

-- Drop policies
DROP POLICY IF EXISTS newsletter_campaign_emails_admin_policy ON newsletter_campaign_emails;
DROP POLICY IF EXISTS newsletter_campaigns_admin_policy ON newsletter_campaigns;
DROP POLICY IF EXISTS newsletter_subscribers_admin_policy ON newsletter_subscribers;

-- Disable RLS
ALTER TABLE IF EXISTS newsletter_campaign_emails DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS newsletter_campaigns DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS newsletter_subscribers DISABLE ROW LEVEL SECURITY;

-- Drop tables (in reverse order of creation due to foreign keys)
DROP TABLE IF EXISTS newsletter_campaign_emails;
DROP TABLE IF EXISTS newsletter_campaigns;
DROP TABLE IF EXISTS newsletter_subscribers;
