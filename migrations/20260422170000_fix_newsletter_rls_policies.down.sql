-- Revert newsletter RLS policies back to admin_role

-- Drop the is_platform_admin() based policies
DROP POLICY IF EXISTS newsletter_subscribers_admin_policy ON newsletter_subscribers;
DROP POLICY IF EXISTS newsletter_campaigns_admin_policy ON newsletter_campaigns;
DROP POLICY IF EXISTS newsletter_campaign_emails_admin_policy ON newsletter_campaign_emails;

-- Recreate policies requiring admin_role
CREATE POLICY newsletter_subscribers_admin_policy ON newsletter_subscribers
    FOR ALL TO admin_role
    USING (true);

CREATE POLICY newsletter_campaigns_admin_policy ON newsletter_campaigns
    FOR ALL TO admin_role
    USING (true);

CREATE POLICY newsletter_campaign_emails_admin_policy ON newsletter_campaign_emails
    FOR ALL TO admin_role
    USING (true);
