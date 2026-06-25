-- FWOS Phase 5: Incidents, Lifecycle, Feature Flags, Data Classification, Certificates, Events

-- Incidents
CREATE TABLE IF NOT EXISTS incidents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  severity TEXT NOT NULL DEFAULT 'medium',
  status TEXT NOT NULL DEFAULT 'open',
  commander_id UUID REFERENCES employees(id),
  project_id UUID REFERENCES projects(id),
  detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  acknowledged_at TIMESTAMPTZ,
  resolved_at TIMESTAMPTZ,
  closed_at TIMESTAMPTZ,
  root_cause TEXT,
  impact TEXT,
  duration_minutes INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT incidents_severity_check CHECK (severity = ANY (ARRAY['low'::text, 'medium'::text, 'high'::text, 'critical'::text])),
  CONSTRAINT incidents_status_check CHECK (status = ANY (ARRAY['open'::text, 'investigating'::text, 'monitoring'::text, 'resolved'::text, 'closed'::text]))
);
CREATE INDEX IF NOT EXISTS idx_incidents_tenant ON incidents(tenant_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_incidents_commander ON incidents(commander_id);

-- Incident Timeline Events
CREATE TABLE IF NOT EXISTS incident_events (
  id BIGSERIAL PRIMARY KEY,
  incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
  author_id UUID NOT NULL REFERENCES employees(id),
  event_type TEXT NOT NULL DEFAULT 'update',
  body TEXT NOT NULL,
  metadata JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT incident_events_type_check CHECK (event_type = ANY (ARRAY['update'::text, 'status_change'::text, 'comment'::text, 'action'::text, 'escalation'::text]))
);
CREATE INDEX IF NOT EXISTS idx_incident_events_incident ON incident_events(incident_id);

-- Incident Responders
CREATE TABLE IF NOT EXISTS incident_responders (
  id BIGSERIAL PRIMARY KEY,
  incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
  employee_id UUID NOT NULL REFERENCES employees(id),
  role TEXT NOT NULL DEFAULT 'responder',
  joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  left_at TIMESTAMPTZ,
  UNIQUE(incident_id, employee_id)
);
CREATE INDEX IF NOT EXISTS idx_incident_responders_incident ON incident_responders(incident_id);

-- Postmortems
CREATE TABLE IF NOT EXISTS postmortems (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  author_id UUID NOT NULL REFERENCES employees(id),
  summary TEXT NOT NULL,
  root_cause TEXT NOT NULL,
  contributing_factors TEXT,
  what_went_well TEXT,
  what_went_wrong TEXT,
  action_items JSONB DEFAULT '[]',
  lessons_learned TEXT,
  status TEXT NOT NULL DEFAULT 'draft',
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT postmortems_status_check CHECK (status = ANY (ARRAY['draft'::text, 'published'::text, 'archived'::text]))
);
CREATE INDEX IF NOT EXISTS idx_postmortems_incident ON postmortems(incident_id);
CREATE INDEX IF NOT EXISTS idx_postmortems_tenant ON postmortems(tenant_id);

-- Employee Lifecycle Events
CREATE TABLE IF NOT EXISTS lifecycle_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  event_type TEXT NOT NULL,
  payload JSONB DEFAULT '{}',
  triggered_by UUID,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT lifecycle_events_type_check CHECK (event_type = ANY (ARRAY['hired'::text, 'onboarded'::text, 'promoted'::text, 'transferred'::text, 'leave_start'::text, 'leave_end'::text, 'offboarding_started'::text, 'terminated'::text, 'reactivated'::text]))
);
CREATE INDEX IF NOT EXISTS idx_lifecycle_events_employee ON lifecycle_events(employee_id);
CREATE INDEX IF NOT EXISTS idx_lifecycle_events_tenant ON lifecycle_events(tenant_id);

-- Lifecycle Workflows
CREATE TABLE IF NOT EXISTS lifecycle_workflows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  trigger_event TEXT NOT NULL,
  steps JSONB NOT NULL DEFAULT '[]',
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_lifecycle_workflows_tenant ON lifecycle_workflows(tenant_id);

-- Lifecycle Workflow Instances
CREATE TABLE IF NOT EXISTS lifecycle_workflow_instances (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_id UUID NOT NULL REFERENCES lifecycle_workflows(id),
  employee_id UUID NOT NULL REFERENCES employees(id),
  tenant_id UUID NOT NULL,
  status TEXT NOT NULL DEFAULT 'in_progress',
  current_step INTEGER DEFAULT 0,
  steps_status JSONB DEFAULT '[]',
  started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT lifecycle_instances_status_check CHECK (status = ANY (ARRAY['in_progress'::text, 'completed'::text, 'cancelled'::text]))
);
CREATE INDEX IF NOT EXISTS idx_lifecycle_instances_employee ON lifecycle_workflow_instances(employee_id);
CREATE INDEX IF NOT EXISTS idx_lifecycle_instances_workflow ON lifecycle_workflow_instances(workflow_id);

-- Feature Flags
CREATE TABLE IF NOT EXISTS feature_flags (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  key TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  flag_type TEXT NOT NULL DEFAULT 'boolean',
  is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  rollout_pct INTEGER DEFAULT 0,
  variants JSONB DEFAULT '{}',
  target_audience JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id, key),
  CONSTRAINT feature_flags_type_check CHECK (flag_type = ANY (ARRAY['boolean'::text, 'percentage'::text, 'variant'::text]))
);
CREATE INDEX IF NOT EXISTS idx_feature_flags_tenant ON feature_flags(tenant_id);
CREATE INDEX IF NOT EXISTS idx_feature_flags_key ON feature_flags(tenant_id, key);

-- Data Classification Labels
CREATE TABLE IF NOT EXISTS data_classifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  resource_type TEXT NOT NULL,
  resource_id UUID NOT NULL,
  classification TEXT NOT NULL DEFAULT 'internal',
  classified_by UUID REFERENCES employees(id),
  reason TEXT,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT data_classifications_level_check CHECK (classification = ANY (ARRAY['public'::text, 'internal'::text, 'confidential'::text, 'restricted'::text, 'founder'::text]))
);
CREATE INDEX IF NOT EXISTS idx_data_classifications_tenant ON data_classifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_data_classifications_resource ON data_classifications(resource_type, resource_id);

-- Employee Certificates (FF-CERT)
CREATE TABLE IF NOT EXISTS employee_certificates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL,
  certificate_serial TEXT UNIQUE NOT NULL,
  certificate_type TEXT NOT NULL DEFAULT 'employee',
  subject TEXT NOT NULL,
  issuer TEXT NOT NULL,
  public_key_pem TEXT NOT NULL,
  fingerprint TEXT NOT NULL,
  device_id TEXT,
  device_name TEXT,
  issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  revoke_reason TEXT,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT employee_certificates_type_check CHECK (certificate_type = ANY (ARRAY['employee'::text, 'manager'::text, 'executive'::text, 'founder'::text])),
  CONSTRAINT employee_certificates_status_check CHECK (status = ANY (ARRAY['active'::text, 'expired'::text, 'revoked'::text]))
);
CREATE INDEX IF NOT EXISTS idx_employee_certificates_employee ON employee_certificates(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_certificates_serial ON employee_certificates(certificate_serial);
CREATE INDEX IF NOT EXISTS idx_employee_certificates_status ON employee_certificates(tenant_id, status);

-- FWOS Event Log
CREATE TABLE IF NOT EXISTS fwos_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  event_type TEXT NOT NULL,
  source TEXT NOT NULL,
  actor_id UUID,
  resource_type TEXT,
  resource_id UUID,
  payload JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fwos_events_tenant ON fwos_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_fwos_events_type ON fwos_events(tenant_id, event_type);
CREATE INDEX IF NOT EXISTS idx_fwos_events_resource ON fwos_events(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_fwos_events_created ON fwos_events(tenant_id, created_at DESC);
