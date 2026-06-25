-- Migration: 20260623170000_fwos_phase2_tables.up.sql
-- Description: FWOS Phase 2 — AI chat, performance management, time tracking, and PTO.
-- This database is SEPARATE from the main functionfly database.
-- user_id and tenant_id columns store UUIDs from the main DB but have NO FK constraints.

-- AI Chat Sessions
CREATE TABLE IF NOT EXISTS ai_chat_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  tenant_id UUID NOT NULL,
  title TEXT DEFAULT 'New Chat',
  context_type TEXT DEFAULT 'general',
  context_reference UUID,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_user ON ai_chat_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_chat_sessions_tenant ON ai_chat_sessions(tenant_id);

-- AI Chat Messages
CREATE TABLE IF NOT EXISTS ai_chat_messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id UUID NOT NULL REFERENCES ai_chat_sessions(id) ON DELETE CASCADE,
  role TEXT NOT NULL,
  content TEXT NOT NULL,
  tokens_used INTEGER DEFAULT 0,
  model TEXT,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT ai_chat_messages_role_check CHECK (role = ANY (ARRAY['user'::text, 'assistant'::text, 'system'::text]))
);
CREATE INDEX IF NOT EXISTS idx_ai_chat_messages_session ON ai_chat_messages(session_id);

-- Performance Goals
CREATE TABLE IF NOT EXISTS performance_goals (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  category TEXT DEFAULT 'professional',
  status TEXT NOT NULL DEFAULT 'not_started',
  priority TEXT DEFAULT 'medium',
  target_date DATE,
  completed_at TIMESTAMPTZ,
  progress_pct INTEGER DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT performance_goals_status_check CHECK (status = ANY (ARRAY['not_started'::text, 'in_progress'::text, 'completed'::text, 'cancelled'::text]))
);
CREATE INDEX IF NOT EXISTS idx_performance_goals_employee ON performance_goals(employee_id);
CREATE INDEX IF NOT EXISTS idx_performance_goals_tenant ON performance_goals(tenant_id);

-- Performance Reviews
CREATE TABLE IF NOT EXISTS performance_reviews (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  reviewer_id UUID NOT NULL REFERENCES employees(id),
  tenant_id UUID NOT NULL,
  review_period TEXT NOT NULL,
  review_type TEXT NOT NULL DEFAULT 'self',
  status TEXT NOT NULL DEFAULT 'draft',
  strengths TEXT,
  areas_for_improvement TEXT,
  overall_rating INTEGER,
  comments TEXT,
  submitted_at TIMESTAMPTZ,
  acknowledged_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT performance_reviews_type_check CHECK (review_type = ANY (ARRAY['self'::text, 'peer'::text, 'manager'::text, 'skip_level'::text])),
  CONSTRAINT performance_reviews_status_check CHECK (status = ANY (ARRAY['draft'::text, 'submitted'::text, 'acknowledged'::text, 'completed'::text])),
  CONSTRAINT performance_reviews_rating_check CHECK (overall_rating >= 1 AND overall_rating <= 5)
);
CREATE INDEX IF NOT EXISTS idx_performance_reviews_employee ON performance_reviews(employee_id);
CREATE INDEX IF NOT EXISTS idx_performance_reviews_reviewer ON performance_reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_performance_reviews_tenant ON performance_reviews(tenant_id);

-- Peer Feedback
CREATE TABLE IF NOT EXISTS peer_feedback (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  review_id UUID REFERENCES performance_reviews(id) ON DELETE CASCADE,
  from_employee_id UUID NOT NULL REFERENCES employees(id),
  to_employee_id UUID NOT NULL REFERENCES employees(id),
  tenant_id UUID NOT NULL,
  feedback_text TEXT NOT NULL,
  rating INTEGER,
  is_anonymous BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT peer_feedback_rating_check CHECK (rating >= 1 AND rating <= 5)
);
CREATE INDEX IF NOT EXISTS idx_peer_feedback_to ON peer_feedback(to_employee_id);
CREATE INDEX IF NOT EXISTS idx_peer_feedback_from ON peer_feedback(from_employee_id);

-- Time Entries
CREATE TABLE IF NOT EXISTS time_entries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
  task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
  date DATE NOT NULL,
  hours NUMERIC(5,2) NOT NULL,
  description TEXT,
  entry_type TEXT NOT NULL DEFAULT 'work',
  is_billable BOOLEAN NOT NULL DEFAULT TRUE,
  approved_by UUID REFERENCES employees(id),
  approved_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT time_entries_type_check CHECK (entry_type = ANY (ARRAY['work'::text, 'meeting'::text, 'training'::text, 'review'::text, 'other'::text]))
);
CREATE INDEX IF NOT EXISTS idx_time_entries_employee ON time_entries(employee_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_date ON time_entries(employee_id, date);
CREATE INDEX IF NOT EXISTS idx_time_entries_project ON time_entries(project_id);

-- PTO Requests
CREATE TABLE IF NOT EXISTS pto_requests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  pto_type TEXT NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  days NUMERIC(4,1) NOT NULL,
  reason TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  approved_by UUID REFERENCES employees(id),
  approved_at TIMESTAMPTZ,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT pto_requests_type_check CHECK (pto_type = ANY (ARRAY['vacation'::text, 'sick'::text, 'personal'::text, 'bereavement'::text, 'jury_duty'::text])),
  CONSTRAINT pto_requests_status_check CHECK (status = ANY (ARRAY['pending'::text, 'approved'::text, 'rejected'::text, 'cancelled'::text]))
);
CREATE INDEX IF NOT EXISTS idx_pto_requests_employee ON pto_requests(employee_id);
CREATE INDEX IF NOT EXISTS idx_pto_requests_tenant ON pto_requests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pto_requests_status ON pto_requests(tenant_id, status);
