-- Migration: Collaborative Sessions
-- Description: Adds support for real-time collaborative playground sessions
-- Users can invite others to jointly edit playground input in real-time

CREATE TABLE IF NOT EXISTS playground_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_key VARCHAR(64) NOT NULL UNIQUE,
    function_id UUID NOT NULL,
    owner_user_id UUID NOT NULL,
    input_state JSONB NOT NULL DEFAULT '{}',
    participants JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    CONSTRAINT fk_function FOREIGN KEY (function_id) REFERENCES registry_functions(id) ON DELETE CASCADE
);

CREATE INDEX idx_playground_sessions_key ON playground_sessions(session_key);
CREATE INDEX idx_playground_sessions_function ON playground_sessions(function_id);
CREATE INDEX idx_playground_sessions_owner ON playground_sessions(owner_user_id);
CREATE INDEX idx_playground_sessions_active ON playground_sessions(is_active) WHERE is_active = true;

CREATE TABLE IF NOT EXISTS playground_session_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL,
    user_id UUID NOT NULL,
    display_name VARCHAR(255),
    cursor_position INTEGER DEFAULT 0,
    selection_range JSONB,
    color VARCHAR(7) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_session FOREIGN KEY (session_id) REFERENCES playground_sessions(id) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT unique_session_participant UNIQUE (session_id, user_id)
);

CREATE INDEX idx_playground_participants_session ON playground_session_participants(session_id);
CREATE INDEX idx_playground_participants_user ON playground_session_participants(user_id);

COMMENT ON TABLE playground_sessions IS 'Real-time collaborative playground editing sessions';
COMMENT ON TABLE playground_session_participants IS 'Participants in collaborative playground sessions';
COMMENT ON COLUMN playground_sessions.session_key IS 'Public shareable key for joining the session';
COMMENT ON COLUMN playground_sessions.input_state IS 'Current JSON input state with operational transforms';
COMMENT ON COLUMN playground_sessions.participants IS 'Array of participant info: [{user_id, display_name, color}]';
COMMENT ON COLUMN playground_session_participants.color IS 'Assigned color for cursor/selection visibility (hex)';