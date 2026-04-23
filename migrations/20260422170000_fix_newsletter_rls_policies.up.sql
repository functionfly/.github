-- Fix newsletter RLS policies to use is_platform_admin() instead of admin_role
-- This matches the pattern used by other tables in the system

-- Drop existing policies that require admin_role
DROP POLICY IF EXISTS newsletter_subscribers_admin_policy ON newsletter_subscribers;
DROP POLICY IF EXISTS newsletter_campaigns_admin_policy ON newsletter_campaigns;
DROP POLICY IF EXISTS newsletter_campaign_emails_admin_policy ON newsletter_campaign_emails;

-- Create admin policy for newsletter_subscribers using is_platform_admin()
-- Admin users can do everything; no tenant isolation needed for newsletter
CREATE POLICY newsletter_subscribers_admin_policy ON newsletter_subscribers
    FOR ALL USING (is_platform_admin());

-- Create admin policy for newsletter_campaigns using is_platform_admin()
CREATE POLICY newsletter_campaigns_admin_policy ON newsletter_campaigns
    FOR ALL USING (is_platform_admin());

-- Create admin policy for newsletter_campaign_emails using is_platform_admin()
CREATE POLICY newsletter_campaign_emails_admin_policy ON newsletter_campaign_emails
    FOR ALL USING (is_platform_admin());
