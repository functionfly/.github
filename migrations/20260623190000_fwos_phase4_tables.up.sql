-- FWOS Phase 4: Team Health, Skills Graph, Reputation, Badges, Living Memory, Mission Control

-- Team Health Metrics
CREATE TABLE IF NOT EXISTS team_health_metrics (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  department_id BIGINT REFERENCES departments(id),
  team_id UUID,
  metric_date DATE NOT NULL,
  workload_score NUMERIC(5,2),
  burnout_risk NUMERIC(5,2),
  velocity_score NUMERIC(5,2),
  collaboration_score NUMERIC(5,2),
  knowledge_sharing_score NUMERIC(5,2),
  pto_utilization_pct NUMERIC(5,2),
  avg_overtime_hours NUMERIC(5,2),
  headcount INTEGER,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_team_health_tenant ON team_health_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_team_health_date ON team_health_metrics(tenant_id, metric_date);
CREATE INDEX IF NOT EXISTS idx_team_health_department ON team_health_metrics(department_id);

-- Skills Graph
CREATE TABLE IF NOT EXISTS skills_graph (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  skill_name TEXT NOT NULL,
  category TEXT,
  total_employees INTEGER DEFAULT 0,
  avg_proficiency NUMERIC(4,2),
  demand_score NUMERIC(5,2),
  supply_score NUMERIC(5,2),
  gap_score NUMERIC(5,2),
  trending BOOLEAN DEFAULT FALSE,
  last_calculated TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, skill_name)
);
CREATE INDEX IF NOT EXISTS idx_skills_graph_tenant ON skills_graph(tenant_id);
CREATE INDEX IF NOT EXISTS idx_skills_graph_gap ON skills_graph(tenant_id, gap_score DESC);

-- Employee Reputation Scores
CREATE TABLE IF NOT EXISTS reputation_scores (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  category TEXT NOT NULL,
  score NUMERIC(7,2) NOT NULL DEFAULT 0,
  rank INTEGER,
  percentile NUMERIC(5,2),
  components JSONB DEFAULT '{}',
  last_calculated TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(employee_id, tenant_id, category)
);
CREATE INDEX IF NOT EXISTS idx_reputation_employee ON reputation_scores(employee_id);
CREATE INDEX IF NOT EXISTS idx_reputation_tenant ON reputation_scores(tenant_id);
CREATE INDEX IF NOT EXISTS idx_reputation_category ON reputation_scores(tenant_id, category, score DESC);

-- Digital Badges
CREATE TABLE IF NOT EXISTS digital_badges (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  slug TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  icon_url TEXT,
  category TEXT DEFAULT 'achievement',
  criteria JSONB DEFAULT '{}',
  points INTEGER DEFAULT 0,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, slug)
);
CREATE INDEX IF NOT EXISTS idx_digital_badges_tenant ON digital_badges(tenant_id);

-- Employee Badge Awards
CREATE TABLE IF NOT EXISTS employee_badges (
  id BIGSERIAL PRIMARY KEY,
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  badge_id UUID NOT NULL REFERENCES digital_badges(id) ON DELETE CASCADE,
  awarded_by UUID,
  awarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(employee_id, badge_id)
);
CREATE INDEX IF NOT EXISTS idx_employee_badges_employee ON employee_badges(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_badges_badge ON employee_badges(badge_id);

-- Living Memory Entries
CREATE TABLE IF NOT EXISTS living_memory (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  author_id UUID NOT NULL REFERENCES employees(id),
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  memory_type TEXT NOT NULL DEFAULT 'note',
  project_id UUID REFERENCES projects(id),
  tags JSONB DEFAULT '[]',
  participants JSONB DEFAULT '[]',
  importance TEXT DEFAULT 'normal',
  searchable_text TEXT,
  view_count INTEGER DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT living_memory_type_check CHECK (memory_type = ANY (ARRAY['meeting'::text, 'decision'::text, 'design'::text, 'lesson'::text, 'discovery'::text, 'note'::text]))
);
CREATE INDEX IF NOT EXISTS idx_living_memory_tenant ON living_memory(tenant_id);
CREATE INDEX IF NOT EXISTS idx_living_memory_type ON living_memory(tenant_id, memory_type);
CREATE INDEX IF NOT EXISTS idx_living_memory_project ON living_memory(project_id);
CREATE INDEX IF NOT EXISTS idx_living_memory_author ON living_memory(author_id);
CREATE INDEX IF NOT EXISTS idx_living_memory_search ON living_memory USING gin(to_tsvector('english', searchable_text));

-- Mission Control Snapshots
CREATE TABLE IF NOT EXISTS mission_control_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  snapshot_date DATE NOT NULL,
  total_employees INTEGER,
  active_employees INTEGER,
  new_hires_30d INTEGER,
  departures_30d INTEGER,
  total_projects INTEGER,
  active_projects INTEGER,
  completed_projects_30d INTEGER,
  total_tasks INTEGER,
  completed_tasks_30d INTEGER,
  avg_task_completion_days NUMERIC(6,2),
  total_learning_hours NUMERIC(8,2),
  avg_skill_proficiency NUMERIC(4,2),
  innovation_grants_submitted INTEGER,
  innovation_grants_funded INTEGER,
  pto_days_used_30d NUMERIC(6,1),
  avg_burnout_risk NUMERIC(5,2),
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, snapshot_date)
);
CREATE INDEX IF NOT EXISTS idx_mission_control_tenant ON mission_control_snapshots(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mission_control_date ON mission_control_snapshots(tenant_id, snapshot_date DESC);
