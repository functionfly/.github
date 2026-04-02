-- Create waitlist table for managing beta access requests
CREATE TABLE IF NOT EXISTS waitlist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    company VARCHAR(255),
    use_case TEXT,
    source VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    invite_code_id UUID,
    invited_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    ip VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_waitlist_email UNIQUE (email),
    CONSTRAINT fk_waitlist_invite_code FOREIGN KEY (invite_code_id) REFERENCES signup_invite_codes(id) ON DELETE SET NULL
);

-- Create index on status for filtering
CREATE INDEX idx_waitlist_status ON waitlist(status);

-- Create index on created_at for sorting
CREATE INDEX idx_waitlist_created_at ON waitlist(created_at DESC);

-- Create index on source for analytics
CREATE INDEX idx_waitlist_source ON waitlist(source);
