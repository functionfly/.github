-- Time Machine: replay, diff, reconcile past function executions
-- Creates 5 tables: replays, replay_items, reconciliations, audit_certificates, schedules

-- 1. Top-level replay job container
CREATE TABLE IF NOT EXISTS time_machine_replays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    window_start TIMESTAMPTZ,
    window_end TIMESTAMPTZ,
    target_version_id UUID,
    target_version VARCHAR(20),
    max_executions INT DEFAULT 1000,
    reconciliation_mode TEXT DEFAULT 'dry_run',
    auto_reconcile BOOLEAN DEFAULT false,
    status TEXT DEFAULT 'pending',
    progress_percent DOUBLE PRECISION DEFAULT 0,
    current_phase TEXT,
    error_message TEXT,
    total_executions_found INT DEFAULT 0,
    total_executions_replayed INT DEFAULT 0,
    total_executions_changed INT DEFAULT 0,
    total_executions_failed INT DEFAULT 0,
    reason TEXT NOT NULL,
    incident_url TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_replays_tenant_id ON time_machine_replays(tenant_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_replays_function_id ON time_machine_replays(function_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_replays_status ON time_machine_replays(status);
CREATE INDEX IF NOT EXISTS idx_time_machine_replays_created_at ON time_machine_replays(created_at DESC);

-- 2. Individual execution replay result
CREATE TABLE IF NOT EXISTS time_machine_replay_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    replay_id UUID NOT NULL REFERENCES time_machine_replays(id) ON DELETE CASCADE,
    original_execution_id UUID,
    original_input JSONB,
    original_output JSONB,
    original_version VARCHAR(20),
    original_duration_ms INT,
    original_timestamp TIMESTAMPTZ,
    original_meg_root_hash TEXT,
    original_certificate_id TEXT,
    new_output JSONB,
    new_duration_ms INT,
    new_meg_root_hash TEXT,
    new_status_code INT,
    output_changed BOOLEAN,
    diff_type TEXT,
    diff_summary TEXT,
    diff_detail JSONB,
    reconciliation_status TEXT DEFAULT 'pending',
    reconciliation_actions JSONB,
    reconciled_at TIMESTAMPTZ,
    replay_error TEXT,
    replay_error_code TEXT,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_replay_items_replay_id ON time_machine_replay_items(replay_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_replay_items_status ON time_machine_replay_items(status);
CREATE INDEX IF NOT EXISTS idx_time_machine_replay_items_diff_type ON time_machine_replay_items(diff_type);
CREATE INDEX IF NOT EXISTS idx_time_machine_replay_items_output_changed ON time_machine_replay_items(output_changed) WHERE output_changed = true;

-- 3. Reconciliation actions log
CREATE TABLE IF NOT EXISTS time_machine_reconciliations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    replay_id UUID NOT NULL REFERENCES time_machine_replays(id) ON DELETE CASCADE,
    replay_item_id UUID NOT NULL REFERENCES time_machine_replay_items(id) ON DELETE CASCADE,
    action_type TEXT NOT NULL,
    target_resource TEXT NOT NULL,
    old_value JSONB,
    new_value JSONB,
    status TEXT DEFAULT 'pending',
    applied_at TIMESTAMPTZ,
    error_message TEXT,
    dry_run BOOLEAN DEFAULT false,
    reversible BOOLEAN DEFAULT true,
    reversal_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_reconciliations_replay_id ON time_machine_reconciliations(replay_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_reconciliations_replay_item_id ON time_machine_reconciliations(replay_item_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_reconciliations_status ON time_machine_reconciliations(status);

-- 4. Compliance-grade audit proof
CREATE TABLE IF NOT EXISTS time_machine_audit_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    replay_id UUID NOT NULL REFERENCES time_machine_replays(id) ON DELETE CASCADE,
    certificate_id TEXT UNIQUE NOT NULL,
    cert_json JSONB NOT NULL,
    cert_hash TEXT NOT NULL,
    previous_cert_hash TEXT,
    merkle_root TEXT,
    signature TEXT,
    compliance_frameworks TEXT[] DEFAULT '{}',
    legal_hold_ref TEXT,
    retention_policy TEXT DEFAULT '7_years',
    anchored BOOLEAN DEFAULT false,
    anchor_chain TEXT,
    anchor_tx_hash TEXT,
    anchored_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_audit_certificates_replay_id ON time_machine_audit_certificates(replay_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_audit_certificates_certificate_id ON time_machine_audit_certificates(certificate_id);

-- 5. Enterprise scheduled replays
CREATE TABLE IF NOT EXISTS time_machine_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    cron_expression TEXT,
    timezone TEXT DEFAULT 'UTC',
    replay_window_hours INT DEFAULT 24,
    target_version_strategy TEXT DEFAULT 'latest',
    pinned_version_id UUID,
    reconciliation_mode TEXT DEFAULT 'dry_run',
    auto_reconcile BOOLEAN DEFAULT false,
    reason_template TEXT DEFAULT 'Scheduled replay',
    enabled BOOLEAN DEFAULT true,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    total_runs INT DEFAULT 0,
    last_replay_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_time_machine_schedules_tenant_id ON time_machine_schedules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_schedules_function_id ON time_machine_schedules(function_id);
CREATE INDEX IF NOT EXISTS idx_time_machine_schedules_enabled ON time_machine_schedules(enabled) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_time_machine_schedules_next_run_at ON time_machine_schedules(next_run_at) WHERE enabled = true;
