-- Feedback System
-- User feedback submissions
CREATE TABLE IF NOT EXISTS feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    user_email VARCHAR(255),
    feedback_type VARCHAR(50) NOT NULL CHECK (feedback_type IN ('bug', 'feature', 'improvement', 'general')),
    subject VARCHAR(500) NOT NULL,
    message TEXT NOT NULL,
    priority VARCHAR(20) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    browser_info TEXT,
    status VARCHAR(20) DEFAULT 'submitted' CHECK (status IN ('submitted', 'in-review', 'resolved', 'closed')),
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Feedback attachments
CREATE TABLE IF NOT EXISTS feedback_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feedback_id UUID NOT NULL REFERENCES feedback(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    s3_key VARCHAR(500) NOT NULL,
    s3_bucket VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_feedback_user_id ON feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_feedback_user_email ON feedback(user_email);
CREATE INDEX IF NOT EXISTS idx_feedback_type ON feedback(feedback_type);
CREATE INDEX IF NOT EXISTS idx_feedback_status ON feedback(status);
CREATE INDEX IF NOT EXISTS idx_feedback_created_at ON feedback(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_feedback_attachments_feedback_id ON feedback_attachments(feedback_id);

-- Updated timestamp trigger for feedback
CREATE TRIGGER update_feedback_updated_at BEFORE UPDATE ON feedback
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();