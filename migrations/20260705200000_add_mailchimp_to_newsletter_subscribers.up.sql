-- Add Mailchimp sync fields to newsletter_subscribers table
ALTER TABLE newsletter_subscribers
  ADD COLUMN IF NOT EXISTS mailchimp_subscriber_id VARCHAR(255),
  ADD COLUMN IF NOT EXISTS mailchimp_list_id VARCHAR(255),
  ADD COLUMN IF NOT EXISTS mailchimp_sync_status VARCHAR(50) DEFAULT 'pending',
  ADD COLUMN IF NOT EXISTS mailchimp_last_synced_at TIMESTAMP,
  ADD COLUMN IF NOT EXISTS email_frequency VARCHAR(20) DEFAULT 'weekly';

CREATE INDEX IF NOT EXISTS idx_newsletter_subscribers_mailchimp_id
  ON newsletter_subscribers(mailchimp_subscriber_id);

CREATE INDEX IF NOT EXISTS idx_newsletter_subscribers_mailchimp_sync_status
  ON newsletter_subscribers(mailchimp_sync_status);
