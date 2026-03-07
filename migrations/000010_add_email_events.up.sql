-- Create email_events table for tracking email delivery, bounces, and engagement
CREATE TABLE IF NOT EXISTS email_events (
    id BIGSERIAL PRIMARY KEY,
    email_id VARCHAR(255) NOT NULL,              -- Resend email ID
    user_id UUID,                                 -- Optional: Associated user
    user_email VARCHAR(255) NOT NULL,             -- Recipient email address
    event_type VARCHAR(50) NOT NULL,              -- Event type: email.sent, email.delivered, email.bounced, email.opened, email.clicked, email.complained
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Event timestamp
    metadata JSONB,                               -- Additional event data (links, bounce details, etc.)
    bounce_reason TEXT,                           -- For bounced emails: reason for bounce
    reviewed BOOLEAN NOT NULL DEFAULT FALSE,      -- Admin review flag for bounces/complaints
    reviewed_by UUID,                             -- Admin who reviewed the event
    reviewed_at TIMESTAMPTZ,                      -- When the event was reviewed
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraint (optional, allows events without user association)
    CONSTRAINT fk_email_events_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_email_events_reviewed_by FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Index on user_id for fast user event lookup
CREATE INDEX IF NOT EXISTS idx_email_events_user_id ON email_events(user_id) WHERE user_id IS NOT NULL;

-- Index on user_email for email-based queries (may not have user_id)
CREATE INDEX IF NOT EXISTS idx_email_events_user_email ON email_events(user_email);

-- Index on event_type for filtering by event type
CREATE INDEX IF NOT EXISTS idx_email_events_event_type ON email_events(event_type);

-- Index on timestamp for time-based queries and analytics
CREATE INDEX IF NOT EXISTS idx_email_events_timestamp ON email_events(timestamp DESC);

-- Index on reviewed flag for admin dashboard (pending review)
CREATE INDEX IF NOT EXISTS idx_email_events_reviewed ON email_events(reviewed) WHERE reviewed = FALSE;

-- Composite index for bounce events needing review
CREATE INDEX IF NOT EXISTS idx_email_events_bounces_pending ON email_events(event_type, reviewed, timestamp DESC) 
WHERE event_type IN ('email.bounced', 'email.complained') AND reviewed = FALSE;

-- Index on email_id for webhook deduplication
CREATE INDEX IF NOT EXISTS idx_email_events_email_id ON email_events(email_id);

-- Comment on table
COMMENT ON TABLE email_events IS 'Tracks email delivery events from Resend webhooks for analytics and bounce management';

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_email_events_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER email_events_updated_at
    BEFORE UPDATE ON email_events
    FOR EACH ROW
    EXECUTE FUNCTION update_email_events_updated_at();
