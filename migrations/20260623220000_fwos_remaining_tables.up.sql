-- 360 Feedback Rounds
CREATE TABLE IF NOT EXISTS feedback_rounds (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  review_period TEXT NOT NULL,
  round_type TEXT NOT NULL DEFAULT '360',
  status TEXT NOT NULL DEFAULT 'draft',
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  questions JSONB DEFAULT '[]',
  created_by UUID NOT NULL REFERENCES employees(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT feedback_rounds_type_check CHECK (round_type = ANY (ARRAY['360'::text, 'upward'::text, 'downward'::text, 'self'::text])),
  CONSTRAINT feedback_rounds_status_check CHECK (status = ANY (ARRAY['draft'::text, 'active'::text, 'collecting'::text, 'completed'::text]))
);
CREATE INDEX IF NOT EXISTS idx_feedback_rounds_tenant ON feedback_rounds(tenant_id);

-- Feedback Round Assignments (who reviews whom)
CREATE TABLE IF NOT EXISTS feedback_round_assignments (
  id BIGSERIAL PRIMARY KEY,
  round_id UUID NOT NULL REFERENCES feedback_rounds(id) ON DELETE CASCADE,
  reviewer_id UUID NOT NULL REFERENCES employees(id),
  reviewee_id UUID NOT NULL REFERENCES employees(id),
  status TEXT NOT NULL DEFAULT 'pending',
  submitted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(round_id, reviewer_id, reviewee_id)
);
CREATE INDEX IF NOT EXISTS idx_feedback_assignments_round ON feedback_round_assignments(round_id);
CREATE INDEX IF NOT EXISTS idx_feedback_assignments_reviewer ON feedback_round_assignments(reviewer_id);

-- Feedback Round Responses
CREATE TABLE IF NOT EXISTS feedback_round_responses (
  id BIGSERIAL PRIMARY KEY,
  assignment_id BIGINT NOT NULL REFERENCES feedback_round_assignments(id) ON DELETE CASCADE,
  question_index INTEGER NOT NULL,
  response_text TEXT,
  response_rating INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(assignment_id, question_index)
);
CREATE INDEX IF NOT EXISTS idx_feedback_responses_assignment ON feedback_round_responses(assignment_id);

-- Goal Cascade (company → team → personal hierarchy)
ALTER TABLE performance_goals ADD COLUMN IF NOT EXISTS parent_goal_id UUID REFERENCES performance_goals(id) ON DELETE SET NULL;
ALTER TABLE performance_goals ADD COLUMN IF NOT EXISTS goal_level TEXT DEFAULT 'personal';
ALTER TABLE performance_goals ADD COLUMN IF NOT EXISTS cascade_visibility TEXT DEFAULT 'private';

-- Document Signatures
CREATE TABLE IF NOT EXISTS document_signatures (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
  signer_id UUID NOT NULL REFERENCES employees(id),
  signer_name TEXT NOT NULL,
  signer_email TEXT,
  signature_data TEXT,
  signed_at TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'pending',
  decline_reason TEXT,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT doc_signatures_status_check CHECK (status = ANY (ARRAY['pending'::text, 'signed'::text, 'declined'::text, 'expired'::text]))
);
CREATE INDEX IF NOT EXISTS idx_document_signatures_document ON document_signatures(document_id);
CREATE INDEX IF NOT EXISTS idx_document_signatures_signer ON document_signatures(signer_id);

-- Employee Certificate Keys (actual PKI)
CREATE TABLE IF NOT EXISTS certificate_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  certificate_id UUID NOT NULL REFERENCES employee_certificates(id) ON DELETE CASCADE,
  private_key_pem TEXT,
  public_key_pem TEXT NOT NULL,
  key_type TEXT NOT NULL DEFAULT 'RSA',
  key_size INTEGER DEFAULT 2048,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_certificate_keys_certificate ON certificate_keys(certificate_id);

-- Wallet Pass Templates
CREATE TABLE IF NOT EXISTS wallet_pass_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  pass_type TEXT NOT NULL,
  platform TEXT NOT NULL,
  template_data JSONB NOT NULL DEFAULT '{}',
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_wallet_templates_tenant ON wallet_pass_templates(tenant_id);

-- Org Chart Import Jobs
CREATE TABLE IF NOT EXISTS org_chart_imports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  uploaded_by UUID NOT NULL REFERENCES employees(id),
  file_name TEXT NOT NULL,
  file_type TEXT NOT NULL DEFAULT 'csv',
  status TEXT NOT NULL DEFAULT 'pending',
  total_rows INTEGER DEFAULT 0,
  processed_rows INTEGER DEFAULT 0,
  error_rows INTEGER DEFAULT 0,
  errors JSONB DEFAULT '[]',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  CONSTRAINT org_imports_status_check CHECK (status = ANY (ARRAY['pending'::text, 'processing'::text, 'completed'::text, 'failed'::text]))
);
CREATE INDEX IF NOT EXISTS idx_org_imports_tenant ON org_chart_imports(tenant_id);

-- Package Registry (internal packages)
CREATE TABLE IF NOT EXISTS package_registry (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  scope TEXT,
  description TEXT,
  registry_type TEXT NOT NULL DEFAULT 'npm',
  latest_version TEXT,
  total_downloads INTEGER DEFAULT 0,
  is_internal BOOLEAN DEFAULT TRUE,
  repository_url TEXT,
  published_by UUID REFERENCES employees(id),
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, name, registry_type)
);
CREATE INDEX IF NOT EXISTS idx_package_registry_tenant ON package_registry(tenant_id);
CREATE INDEX IF NOT EXISTS idx_package_registry_name ON package_registry(tenant_id, name);

-- Package Versions
CREATE TABLE IF NOT EXISTS package_versions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  package_id UUID NOT NULL REFERENCES package_registry(id) ON DELETE CASCADE,
  version TEXT NOT NULL,
  description TEXT,
  dependencies JSONB DEFAULT '{}',
  published_by UUID REFERENCES employees(id),
  downloads INTEGER DEFAULT 0,
  tarball_url TEXT,
  published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(package_id, version)
);
CREATE INDEX IF NOT EXISTS idx_package_versions_package ON package_versions(package_id);
