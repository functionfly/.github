-- Achievement Definitions (pre-defined badges)
CREATE TABLE IF NOT EXISTS achievement_definitions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  slug TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  icon TEXT,
  category TEXT NOT NULL DEFAULT 'achievement',
  criteria_type TEXT NOT NULL,
  criteria_threshold INTEGER NOT NULL,
  points INTEGER NOT NULL DEFAULT 0,
  tier INTEGER NOT NULL DEFAULT 1,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, slug)
);
CREATE INDEX IF NOT EXISTS idx_achievement_definitions_tenant ON achievement_definitions(tenant_id);

-- Achievement Progress (tracks progress toward achievements)
CREATE TABLE IF NOT EXISTS achievement_progress (
  id BIGSERIAL PRIMARY KEY,
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  achievement_id UUID NOT NULL REFERENCES achievement_definitions(id) ON DELETE CASCADE,
  current_value INTEGER NOT NULL DEFAULT 0,
  awarded BOOLEAN NOT NULL DEFAULT FALSE,
  awarded_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(employee_id, achievement_id)
);
CREATE INDEX IF NOT EXISTS idx_achievement_progress_employee ON achievement_progress(employee_id);

-- Career Timeline Events
CREATE TABLE IF NOT EXISTS career_timeline_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  event_type TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  metadata JSONB DEFAULT '{}',
  event_date DATE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT career_timeline_type_check CHECK (event_type = ANY (ARRAY['joined'::text, 'promoted'::text, 'transferred'::text, 'project_completed'::text, 'certification_earned'::text, 'achievement_unlocked'::text, 'award_received'::text]))
);
CREATE INDEX IF NOT EXISTS idx_career_timeline_employee ON career_timeline_events(employee_id);
CREATE INDEX IF NOT EXISTS idx_career_timeline_date ON career_timeline_events(employee_id, event_date DESC);

-- Reputation History (for trend charts)
CREATE TABLE IF NOT EXISTS reputation_history (
  id BIGSERIAL PRIMARY KEY,
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  category TEXT NOT NULL,
  score NUMERIC(7,2) NOT NULL,
  recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reputation_history_employee ON reputation_history(employee_id);
CREATE INDEX IF NOT EXISTS idx_reputation_history_category ON reputation_history(employee_id, category, recorded_at DESC);

-- Add identity columns to employees
ALTER TABLE employees ADD COLUMN IF NOT EXISTS clearance_level_num INTEGER DEFAULT 1;
ALTER TABLE employees ADD COLUMN IF NOT EXISTS identity_signature TEXT;
ALTER TABLE employees ADD COLUMN IF NOT EXISTS reputation_total INTEGER DEFAULT 0;
ALTER TABLE employees ADD COLUMN IF NOT EXISTS trust_score NUMERIC(5,2) DEFAULT 0;
