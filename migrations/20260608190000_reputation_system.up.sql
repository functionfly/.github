-- Migration: Add reputation farming alerts and trust score weights config tables
-- Created: 20260608190000

-- Reputation Farming Alerts table
CREATE TABLE IF NOT EXISTS reputation_farming_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    description TEXT,
    affected_functions JSONB DEFAULT '[]'::jsonb,
    affected_users JSONB DEFAULT '[]'::jsonb,
    severity VARCHAR(20) DEFAULT 'low',
    status VARCHAR(20) DEFAULT 'open',
    detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID,
    notes TEXT,
    details JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_reputation_alerts_status ON reputation_farming_alerts(status);
CREATE INDEX IF NOT EXISTS idx_reputation_alerts_severity ON reputation_farming_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_reputation_alerts_detected_at ON reputation_farming_alerts(detected_at DESC);

-- Trust Score Weights Config table
CREATE TABLE IF NOT EXISTS trust_score_weights_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    reliability NUMERIC(5, 4) DEFAULT 0.30,
    latency NUMERIC(5, 4) DEFAULT 0.20,
    error_rate NUMERIC(5, 4) DEFAULT 0.20,
    user_rating NUMERIC(5, 4) DEFAULT 0.15,
    verification NUMERIC(5, 4) DEFAULT 0.15,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trust_weights_active ON trust_score_weights_config(is_active) WHERE is_active = TRUE;

-- Reputation Events table (audit trail)
CREATE TABLE IF NOT EXISTS reputation_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    score_change INT DEFAULT 0,
    component VARCHAR(50),
    reference_id UUID,
    description TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reputation_events_user_id ON reputation_events(user_id);
CREATE INDEX IF NOT EXISTS idx_reputation_events_created_at ON reputation_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reputation_events_type ON reputation_events(event_type);

-- Add foreign key constraint for reputation_profiles if not exists
-- Note: This assumes users table already exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'reputation_profiles_user_id_fkey'
    ) THEN
        ALTER TABLE reputation_profiles
        ADD CONSTRAINT reputation_profiles_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Add foreign key for resolved_by in reputation_farming_alerts
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'reputation_farming_alerts_resolved_by_fkey'
    ) THEN
        ALTER TABLE reputation_farming_alerts
        ADD CONSTRAINT reputation_farming_alerts_resolved_by_fkey
        FOREIGN KEY (resolved_by) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Add foreign key for user_id in reputation_events
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'reputation_events_user_id_fkey'
    ) THEN
        ALTER TABLE reputation_events
        ADD CONSTRAINT reputation_events_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;
