-- Migration: Support System Tables
-- Description: AI + Human Co-Pilot Support System

-- Create enum types
DO $$ BEGIN
    CREATE TYPE support_conversation_type_enum AS ENUM (
        'support_ai',
        'support_human',
        'support_emergency'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE support_status_enum AS ENUM (
        'active',
        'pending',
        'resolved',
        'escalated'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE support_priority_enum AS ENUM (
        'low',
        'normal',
        'high',
        'critical'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE support_author_type_enum AS ENUM (
        'user',
        'ai',
        'staff',
        'system'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE support_message_type_enum AS ENUM (
        'message',
        'context',
        'escalation',
        'resolution',
        'ai_response',
        'system'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Support Conversations table
CREATE TABLE IF NOT EXISTS support_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type support_conversation_type_enum NOT NULL DEFAULT 'support_ai',
    status support_status_enum NOT NULL DEFAULT 'active',
    priority support_priority_enum NOT NULL DEFAULT 'normal',
    title TEXT,

    -- Function reference (stored as JSON)
    function_ref JSONB DEFAULT '{}',

    -- Deployment information
    deployment_id UUID,
    deployment_logs TEXT,
    deployment_error TEXT,

    -- AI handling
    ai_handled BOOLEAN DEFAULT false,
    ai_attempts INTEGER DEFAULT 0,

    -- Human escalation
    staff_id UUID,
    staff_joined_at TIMESTAMPTZ,

    -- Resolution
    resolved_at TIMESTAMPTZ,
    resolved_by_id UUID,
    resolution_note TEXT,

    -- Emergency handling
    is_emergency BOOLEAN DEFAULT false,
    emergency_code TEXT,

    -- Metadata
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index for user conversations
CREATE INDEX IF NOT EXISTS idx_support_conversations_user_id ON support_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_support_conversations_status ON support_conversations(status);
CREATE INDEX IF NOT EXISTS idx_support_conversations_created_at ON support_conversations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_support_conversations_emergency ON support_conversations(is_emergency) WHERE is_emergency = true;

-- Support Messages table
CREATE TABLE IF NOT EXISTS support_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES support_conversations(id) ON DELETE CASCADE,
    author_id UUID NOT NULL,
    author_type support_author_type_enum NOT NULL DEFAULT 'user',
    message_type support_message_type_enum NOT NULL DEFAULT 'message',
    content TEXT NOT NULL DEFAULT '',

    -- AI-specific fields
    ai_confidence FLOAT,
    ai_model TEXT,

    -- Context embedding
    embeddings JSONB DEFAULT '{}',

    -- Attachments (logs, screenshots, etc.)
    attachments JSONB DEFAULT '[]',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_support_messages_conversation_id ON support_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_support_messages_created_at ON support_messages(created_at ASC);

-- Staff Availability table
CREATE TABLE IF NOT EXISTS staff_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    staff_id UUID NOT NULL UNIQUE,
    is_online BOOLEAN DEFAULT false,
    last_seen TIMESTAMPTZ,
    max_chats INTEGER DEFAULT 5,
    active_chats INTEGER DEFAULT 0,
    can_accept BOOLEAN DEFAULT true,
    specialties JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_staff_availability_online ON staff_availability(is_online) WHERE is_online = true;

-- Support Conversation Participants table
CREATE TABLE IF NOT EXISTS support_conversation_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES support_conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role TEXT NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL,
    left_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_support_participants_conversation_id ON support_conversation_participants(conversation_id);
CREATE INDEX IF NOT EXISTS idx_support_participants_user_id ON support_conversation_participants(user_id);

-- Emergency Fix Requests table
CREATE TABLE IF NOT EXISTS emergency_fix_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES support_conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    function_id UUID NOT NULL,
    reason TEXT,

    -- Status tracking
    status TEXT NOT NULL DEFAULT 'pending',
    staff_id UUID,
    staff_accepted_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,

    -- What was done
    fix_description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_emergency_requests_conversation_id ON emergency_fix_requests(conversation_id);
CREATE INDEX IF NOT EXISTS idx_emergency_requests_status ON emergency_fix_requests(status);
CREATE INDEX IF NOT EXISTS idx_emergency_requests_created_at ON emergency_fix_requests(created_at ASC);

-- Updated at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
DROP TRIGGER IF EXISTS update_support_conversations_updated_at ON support_conversations;
CREATE TRIGGER update_support_conversations_updated_at
    BEFORE UPDATE ON support_conversations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_staff_availability_updated_at ON staff_availability;
CREATE TRIGGER update_staff_availability_updated_at
    BEFORE UPDATE ON staff_availability
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_emergency_requests_updated_at ON emergency_fix_requests;
CREATE TRIGGER update_emergency_requests_updated_at
    BEFORE UPDATE ON emergency_fix_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE support_conversations IS 'AI + Human Co-Pilot support conversation sessions';
COMMENT ON TABLE support_messages IS 'Messages within support conversations';
COMMENT ON TABLE staff_availability IS 'Staff online/offline status for support routing';
COMMENT ON TABLE support_conversation_participants IS 'Tracks who is in each support conversation';
COMMENT ON TABLE emergency_fix_requests IS 'Emergency fix button activations for production failures';
