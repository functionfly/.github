-- Innovation Grants
CREATE TABLE IF NOT EXISTS innovation_grants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  proposer_id UUID NOT NULL REFERENCES employees(id),
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  category TEXT DEFAULT 'technical',
  requested_amount_cents BIGINT,
  status TEXT NOT NULL DEFAULT 'draft',
  feasibility_score NUMERIC(5,2),
  votes_for INTEGER DEFAULT 0,
  votes_against INTEGER DEFAULT 0,
  reviewed_by UUID REFERENCES employees(id),
  reviewed_at TIMESTAMPTZ,
  rejection_reason TEXT,
  funded_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT innovation_grants_status_check CHECK (status = ANY (ARRAY['draft'::text, 'submitted'::text, 'under_review'::text, 'approved'::text, 'rejected'::text, 'funded'::text, 'completed'::text]))
);
CREATE INDEX IF NOT EXISTS idx_innovation_grants_tenant ON innovation_grants(tenant_id);
CREATE INDEX IF NOT EXISTS idx_innovation_grants_proposer ON innovation_grants(proposer_id);
CREATE INDEX IF NOT EXISTS idx_innovation_grants_status ON innovation_grants(tenant_id, status);

-- Innovation Grant Votes
CREATE TABLE IF NOT EXISTS innovation_grant_votes (
  id BIGSERIAL PRIMARY KEY,
  grant_id UUID NOT NULL REFERENCES innovation_grants(id) ON DELETE CASCADE,
  voter_id UUID NOT NULL REFERENCES employees(id),
  vote BOOLEAN NOT NULL,
  comment TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(grant_id, voter_id)
);
CREATE INDEX IF NOT EXISTS idx_innovation_grant_votes_grant ON innovation_grant_votes(grant_id);

-- Talent Marketplace Opportunities
CREATE TABLE IF NOT EXISTS marketplace_opportunities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  posted_by UUID NOT NULL REFERENCES employees(id),
  department_id BIGINT REFERENCES departments(id),
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  opportunity_type TEXT NOT NULL DEFAULT 'project',
  skills_required JSONB DEFAULT '[]',
  hours_per_week NUMERIC(4,1),
  duration_weeks INTEGER,
  is_remote BOOLEAN DEFAULT TRUE,
  status TEXT NOT NULL DEFAULT 'open',
  max_applicants INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT marketplace_opportunities_type_check CHECK (opportunity_type = ANY (ARRAY['project'::text, 'gig'::text, 'mentorship'::text, 'hackathon'::text])),
  CONSTRAINT marketplace_opportunities_status_check CHECK (status = ANY (ARRAY['open'::text, 'filled'::text, 'closed'::text, 'cancelled'::text]))
);
CREATE INDEX IF NOT EXISTS idx_marketplace_opportunities_tenant ON marketplace_opportunities(tenant_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_opportunities_status ON marketplace_opportunities(tenant_id, status);

-- Marketplace Applications
CREATE TABLE IF NOT EXISTS marketplace_applications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  opportunity_id UUID NOT NULL REFERENCES marketplace_opportunities(id) ON DELETE CASCADE,
  applicant_id UUID NOT NULL REFERENCES employees(id),
  message TEXT,
  status TEXT NOT NULL DEFAULT 'pending',
  reviewed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(opportunity_id, applicant_id),
  CONSTRAINT marketplace_applications_status_check CHECK (status = ANY (ARRAY['pending'::text, 'accepted'::text, 'rejected'::text, 'withdrawn'::text]))
);
CREATE INDEX IF NOT EXISTS idx_marketplace_applications_opportunity ON marketplace_applications(opportunity_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_applications_applicant ON marketplace_applications(applicant_id);

-- Career Paths
CREATE TABLE IF NOT EXISTS career_paths (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  title TEXT NOT NULL,
  track TEXT NOT NULL,
  level INTEGER NOT NULL,
  description TEXT,
  requirements JSONB DEFAULT '{}',
  salary_range_min_cents BIGINT,
  salary_range_max_cents BIGINT,
  next_path_id UUID REFERENCES career_paths(id),
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_career_paths_tenant ON career_paths(tenant_id);
CREATE INDEX IF NOT EXISTS idx_career_paths_track ON career_paths(tenant_id, track);

-- Employee Career Progress
CREATE TABLE IF NOT EXISTS employee_career_progress (
  id BIGSERIAL PRIMARY KEY,
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  career_path_id UUID NOT NULL REFERENCES career_paths(id),
  status TEXT NOT NULL DEFAULT 'current',
  gap_analysis JSONB DEFAULT '{}',
  started_at TIMESTAMPTZ,
  target_date DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_employee_career_progress_employee ON employee_career_progress(employee_id);

-- Mentorship Matches
CREATE TABLE IF NOT EXISTS mentorship_matches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  mentor_id UUID NOT NULL REFERENCES employees(id),
  mentee_id UUID NOT NULL REFERENCES employees(id),
  focus_area TEXT,
  status TEXT NOT NULL DEFAULT 'active',
  started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ended_at TIMESTAMPTZ,
  meeting_frequency TEXT DEFAULT 'biweekly',
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT mentorship_status_check CHECK (status = ANY (ARRAY['active'::text, 'paused'::text, 'completed'::text, 'cancelled'::text]))
);
CREATE INDEX IF NOT EXISTS idx_mentorship_matches_mentor ON mentorship_matches(mentor_id);
CREATE INDEX IF NOT EXISTS idx_mentorship_matches_mentee ON mentorship_matches(mentee_id);

-- Documents
CREATE TABLE IF NOT EXISTS documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  author_id UUID NOT NULL REFERENCES employees(id),
  title TEXT NOT NULL,
  body TEXT,
  doc_type TEXT NOT NULL DEFAULT 'note',
  category TEXT,
  tags JSONB DEFAULT '[]',
  is_template BOOLEAN DEFAULT FALSE,
  status TEXT NOT NULL DEFAULT 'draft',
  view_count INTEGER DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT documents_type_check CHECK (doc_type = ANY (ARRAY['note'::text, 'policy'::text, 'template'::text, 'meeting_notes'::text, 'design_doc'::text])),
  CONSTRAINT documents_status_check CHECK (status = ANY (ARRAY['draft'::text, 'published'::text, 'archived'::text]))
);
CREATE INDEX IF NOT EXISTS idx_documents_tenant ON documents(tenant_id);
CREATE INDEX IF NOT EXISTS idx_documents_author ON documents(author_id);
CREATE INDEX IF NOT EXISTS idx_documents_type ON documents(tenant_id, doc_type);

-- Document Shares
CREATE TABLE IF NOT EXISTS document_shares (
  id BIGSERIAL PRIMARY KEY,
  document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  shared_with UUID NOT NULL REFERENCES employees(id),
  permission TEXT NOT NULL DEFAULT 'read',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(document_id, shared_with)
);
CREATE INDEX IF NOT EXISTS idx_document_shares_document ON document_shares(document_id);
