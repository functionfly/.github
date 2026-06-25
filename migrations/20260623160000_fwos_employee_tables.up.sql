-- Migration: 20260623160000_fwos_employee_tables.up.sql
-- Description: FWOS (FunctionFly Workspace OS) employee operating system tables.
-- This database is SEPARATE from the main functionfly database.
-- user_id and tenant_id columns store UUIDs from the main DB but have NO FK constraints
-- since those tables live in a different database.

-- ============================================================================
-- departments
-- Organizational hierarchy. Self-referencing for nested departments.
-- ============================================================================
CREATE TABLE IF NOT EXISTS departments (
  id          BIGSERIAL PRIMARY KEY,
  tenant_id   UUID NOT NULL,
  name        TEXT NOT NULL,
  slug        TEXT NOT NULL,
  description TEXT,
  parent_id   BIGINT REFERENCES departments(id) ON DELETE SET NULL,
  head_id     UUID,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_departments_tenant ON departments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_departments_parent ON departments(parent_id);

-- ============================================================================
-- employees
-- 1:1 extension of users table (user_id references main DB, no FK).
-- ============================================================================
CREATE TABLE IF NOT EXISTS employees (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID NOT NULL UNIQUE,
  tenant_id       UUID NOT NULL,
  employee_number TEXT UNIQUE NOT NULL,
  ffid            TEXT UNIQUE NOT NULL,
  department_id   BIGINT REFERENCES departments(id) ON DELETE SET NULL,
  manager_id      UUID REFERENCES employees(id) ON DELETE SET NULL,
  hire_date       DATE,
  employment_type TEXT NOT NULL DEFAULT 'full_time',
  clearance_level TEXT NOT NULL DEFAULT 'standard',
  work_location   TEXT,
  office_location TEXT,
  timezone        TEXT,
  bio             TEXT,
  pronouns        TEXT,
  emergency_contact JSONB,
  status          TEXT NOT NULL DEFAULT 'active',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT employees_employment_type_check
    CHECK (employment_type = ANY (ARRAY['full_time'::text, 'part_time'::text, 'contractor'::text, 'intern'::text])),
  CONSTRAINT employees_clearance_check
    CHECK (clearance_level = ANY (ARRAY['standard'::text, 'elevated'::text, 'confidential'::text, 'top_secret'::text])),
  CONSTRAINT employees_status_check
    CHECK (status = ANY (ARRAY['active'::text, 'on_leave'::text, 'terminated'::text, 'suspended'::text]))
);

CREATE INDEX IF NOT EXISTS idx_employees_tenant ON employees(tenant_id);
CREATE INDEX IF NOT EXISTS idx_employees_manager ON employees(manager_id);
CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department_id);
CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(tenant_id, status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_employees_ffid ON employees(ffid);

-- Add FK from departments.head_id → employees now that employees exists
ALTER TABLE departments ADD CONSTRAINT departments_head_id_fkey
  FOREIGN KEY (head_id) REFERENCES employees(id) ON DELETE SET NULL;

-- ============================================================================
-- employee_departments
-- ============================================================================
CREATE TABLE IF NOT EXISTS employee_departments (
  id            BIGSERIAL PRIMARY KEY,
  employee_id   UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  department_id BIGINT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
  role_in_dept  TEXT NOT NULL DEFAULT 'member',
  is_primary    BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(employee_id, department_id)
);

CREATE INDEX IF NOT EXISTS idx_employee_departments_employee ON employee_departments(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_departments_department ON employee_departments(department_id);

-- ============================================================================
-- employee_skills
-- ============================================================================
CREATE TABLE IF NOT EXISTS employee_skills (
  id           BIGSERIAL PRIMARY KEY,
  employee_id  UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  skill_name   TEXT NOT NULL,
  category     TEXT,
  proficiency  TEXT NOT NULL DEFAULT 'intermediate',
  years_exp    NUMERIC(4,1),
  endorsements INTEGER NOT NULL DEFAULT 0,
  verified     BOOLEAN NOT NULL DEFAULT FALSE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(employee_id, skill_name),
  CONSTRAINT employee_skills_proficiency_check
    CHECK (proficiency = ANY (ARRAY['beginner'::text, 'intermediate'::text, 'advanced'::text, 'expert'::text]))
);

CREATE INDEX IF NOT EXISTS idx_employee_skills_employee ON employee_skills(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_skills_name ON employee_skills(skill_name);

-- ============================================================================
-- employee_certifications
-- ============================================================================
CREATE TABLE IF NOT EXISTS employee_certifications (
  id              BIGSERIAL PRIMARY KEY,
  employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,
  issuer          TEXT NOT NULL,
  credential_id   TEXT,
  credential_url  TEXT,
  issued_date     DATE,
  expiry_date     DATE,
  verified        BOOLEAN NOT NULL DEFAULT FALSE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_employee_certifications_employee ON employee_certifications(employee_id);

-- ============================================================================
-- employee_achievements
-- ============================================================================
CREATE TABLE IF NOT EXISTS employee_achievements (
  id           BIGSERIAL PRIMARY KEY,
  employee_id  UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  title        TEXT NOT NULL,
  description  TEXT,
  type         TEXT NOT NULL DEFAULT 'recognition',
  awarded_by   UUID,
  points       INTEGER DEFAULT 0,
  badge_url    TEXT,
  awarded_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT employee_achievements_type_check
    CHECK (type = ANY (ARRAY['recognition'::text, 'award'::text, 'milestone'::text, 'peer_kudos'::text]))
);

CREATE INDEX IF NOT EXISTS idx_employee_achievements_employee ON employee_achievements(employee_id);

-- ============================================================================
-- projects
-- ============================================================================
CREATE TABLE IF NOT EXISTS projects (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID NOT NULL,
  name         TEXT NOT NULL,
  slug         TEXT NOT NULL,
  description  TEXT,
  status       TEXT NOT NULL DEFAULT 'active',
  priority     TEXT NOT NULL DEFAULT 'medium',
  owner_id     UUID NOT NULL REFERENCES employees(id),
  start_date   DATE,
  target_date  DATE,
  tags         JSONB DEFAULT '[]',
  metadata     JSONB DEFAULT '{}',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, slug),
  CONSTRAINT projects_status_check
    CHECK (status = ANY (ARRAY['active'::text, 'paused'::text, 'completed'::text, 'archived'::text])),
  CONSTRAINT projects_priority_check
    CHECK (priority = ANY (ARRAY['low'::text, 'medium'::text, 'high'::text, 'critical'::text]))
);

CREATE INDEX IF NOT EXISTS idx_projects_tenant ON projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(tenant_id, status);

-- ============================================================================
-- tasks
-- ============================================================================
CREATE TABLE IF NOT EXISTS tasks (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  tenant_id       UUID NOT NULL,
  title           TEXT NOT NULL,
  description     TEXT,
  status          TEXT NOT NULL DEFAULT 'todo',
  priority        TEXT NOT NULL DEFAULT 'medium',
  assignee_id     UUID REFERENCES employees(id) ON DELETE SET NULL,
  reporter_id     UUID NOT NULL REFERENCES employees(id),
  parent_id       UUID REFERENCES tasks(id) ON DELETE SET NULL,
  due_date        DATE,
  estimated_hours NUMERIC(6,1),
  actual_hours    NUMERIC(6,1),
  tags            JSONB DEFAULT '[]',
  position        INTEGER DEFAULT 0,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT tasks_status_check
    CHECK (status = ANY (ARRAY['todo'::text, 'in_progress'::text, 'in_review'::text, 'done'::text, 'blocked'::text])),
  CONSTRAINT tasks_priority_check
    CHECK (priority = ANY (ARRAY['low'::text, 'medium'::text, 'high'::text, 'critical'::text]))
);

CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee ON tasks(assignee_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant_status ON tasks(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_due ON tasks(due_date) WHERE status NOT IN ('done', 'archived');
CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id);

-- ============================================================================
-- task_comments
-- ============================================================================
CREATE TABLE IF NOT EXISTS task_comments (
  id         BIGSERIAL PRIMARY KEY,
  task_id    UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  author_id  UUID NOT NULL REFERENCES employees(id),
  body       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_comments_task ON task_comments(task_id);

-- ============================================================================
-- learning_courses
-- ============================================================================
CREATE TABLE IF NOT EXISTS learning_courses (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      UUID NOT NULL,
  title          TEXT NOT NULL,
  description    TEXT,
  category       TEXT,
  difficulty     TEXT DEFAULT 'beginner',
  duration_min   INTEGER,
  content_url    TEXT,
  thumbnail_url  TEXT,
  is_mandatory   BOOLEAN NOT NULL DEFAULT FALSE,
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_learning_courses_tenant ON learning_courses(tenant_id);
CREATE INDEX IF NOT EXISTS idx_learning_courses_category ON learning_courses(tenant_id, category) WHERE is_active = TRUE;

-- ============================================================================
-- employee_learning
-- ============================================================================
CREATE TABLE IF NOT EXISTS employee_learning (
  id            BIGSERIAL PRIMARY KEY,
  employee_id   UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  course_id     UUID NOT NULL REFERENCES learning_courses(id) ON DELETE CASCADE,
  status        TEXT NOT NULL DEFAULT 'not_started',
  progress_pct  INTEGER DEFAULT 0,
  started_at    TIMESTAMPTZ,
  completed_at  TIMESTAMPTZ,
  score         NUMERIC(5,2),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(employee_id, course_id),
  CONSTRAINT employee_learning_status_check
    CHECK (status = ANY (ARRAY['not_started'::text, 'in_progress'::text, 'completed'::text, 'skipped'::text]))
);

CREATE INDEX IF NOT EXISTS idx_employee_learning_employee ON employee_learning(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_learning_course ON employee_learning(course_id);

-- ============================================================================
-- knowledge_articles
-- ============================================================================
CREATE TABLE IF NOT EXISTS knowledge_articles (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID NOT NULL,
  title        TEXT NOT NULL,
  slug         TEXT NOT NULL,
  body         TEXT NOT NULL,
  category     TEXT,
  tags         JSONB DEFAULT '[]',
  author_id    UUID NOT NULL REFERENCES employees(id),
  status       TEXT NOT NULL DEFAULT 'draft',
  view_count   INTEGER DEFAULT 0,
  published_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, slug),
  CONSTRAINT knowledge_articles_status_check
    CHECK (status = ANY (ARRAY['draft'::text, 'published'::text, 'archived'::text]))
);

CREATE INDEX IF NOT EXISTS idx_knowledge_articles_tenant ON knowledge_articles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_status ON knowledge_articles(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_author ON knowledge_articles(author_id);

-- ============================================================================
-- compensation_records
-- ============================================================================
CREATE TABLE IF NOT EXISTS compensation_records (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id       UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id         UUID NOT NULL,
  base_salary_cents BIGINT NOT NULL,
  currency          TEXT NOT NULL DEFAULT 'USD',
  pay_frequency     TEXT NOT NULL DEFAULT 'biweekly',
  effective_date    DATE NOT NULL,
  end_date          DATE,
  review_date       DATE,
  notes             TEXT,
  created_by        UUID NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT compensation_pay_frequency_check
    CHECK (pay_frequency = ANY (ARRAY['weekly'::text, 'biweekly'::text, 'monthly'::text]))
);

CREATE INDEX IF NOT EXISTS idx_compensation_employee ON compensation_records(employee_id);
CREATE INDEX IF NOT EXISTS idx_compensation_tenant ON compensation_records(tenant_id);

-- ============================================================================
-- equity_grants
-- ============================================================================
CREATE TABLE IF NOT EXISTS equity_grants (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id        UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id          UUID NOT NULL,
  grant_type         TEXT NOT NULL,
  shares             INTEGER NOT NULL,
  strike_price_cents BIGINT,
  vesting_start      DATE NOT NULL,
  vesting_end        DATE NOT NULL,
  cliff_date         DATE,
  vested_shares      INTEGER DEFAULT 0,
  status             TEXT NOT NULL DEFAULT 'active',
  grant_date         DATE NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT equity_grants_type_check
    CHECK (grant_type = ANY (ARRAY['iso'::text, 'nso'::text, 'rsu'::text, 'phantom'::text])),
  CONSTRAINT equity_grants_status_check
    CHECK (status = ANY (ARRAY['active'::text, 'vested'::text, 'cancelled'::text, 'exercised'::text]))
);

CREATE INDEX IF NOT EXISTS idx_equity_grants_employee ON equity_grants(employee_id);

-- ============================================================================
-- compensation_access_log
-- ============================================================================
CREATE TABLE IF NOT EXISTS compensation_access_log (
  id          BIGSERIAL PRIMARY KEY,
  accessor_id UUID NOT NULL,
  target_id   UUID NOT NULL REFERENCES employees(id),
  action      TEXT NOT NULL,
  ip_address  INET,
  user_agent  TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT compensation_access_action_check
    CHECK (action = ANY (ARRAY['view'::text, 'export'::text, 'modify'::text]))
);

CREATE INDEX IF NOT EXISTS idx_compensation_access_log_target ON compensation_access_log(target_id);
CREATE INDEX IF NOT EXISTS idx_compensation_access_log_accessor ON compensation_access_log(accessor_id);

-- ============================================================================
-- fwos_notifications
-- ============================================================================
CREATE TABLE IF NOT EXISTS fwos_notifications (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL,
  tenant_id  UUID NOT NULL,
  type       TEXT NOT NULL,
  title      TEXT NOT NULL,
  body       TEXT,
  action_url TEXT,
  is_read    BOOLEAN NOT NULL DEFAULT FALSE,
  read_at    TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fwos_notifications_user ON fwos_notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_fwos_notifications_tenant ON fwos_notifications(tenant_id);
