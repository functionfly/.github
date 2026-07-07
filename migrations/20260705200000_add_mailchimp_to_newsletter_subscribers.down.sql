-- Remove Mailchimp sync fields from newsletter_subscribers table
ALTER TABLE newsletter_subscribers
  DROP COLUMN IF EXISTS mailchimp_subscriber_id,
  DROP COLUMN IF EXISTS mailchimp_list_id,
  DROP COLUMN IF EXISTS mailchimp_sync_status,
  DROP COLUMN IF EXISTS mailchimp_last_synced_at,
  DROP COLUMN IF EXISTS email_frequency;

DROP INDEX IF EXISTS idx_newsletter_subscribers_mailchimp_id;
DROP INDEX IF EXISTS idx_newsletter_subscribers_mailchimp_sync_status;
