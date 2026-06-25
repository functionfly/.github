-- Prism Runtime: cells, heartbeats, execution results, and capabilities tables.
-- These persist state from the Prism WASM runtime received via NATS.

CREATE TABLE IF NOT EXISTS prism_cells (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cell_id         TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    runtime         TEXT NOT NULL DEFAULT 'prism',
    status          TEXT NOT NULL DEFAULT 'registered',
    capabilities    TEXT[] NOT NULL DEFAULT '{}',
    metadata        JSONB NOT NULL DEFAULT '{}',
    tenant_id       UUID,
    registered_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_heartbeat  TIMESTAMPTZ,
    terminated_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prism_cells_cell_id ON prism_cells(cell_id);
CREATE INDEX IF NOT EXISTS idx_prism_cells_status ON prism_cells(status);
CREATE INDEX IF NOT EXISTS idx_prism_cells_tenant ON prism_cells(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_prism_cells_last_heartbeat ON prism_cells(last_heartbeat DESC);

CREATE TABLE IF NOT EXISTS prism_heartbeats (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cell_id         TEXT NOT NULL,
    status          TEXT NOT NULL,
    active_executions INTEGER NOT NULL DEFAULT 0,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prism_heartbeats_cell_id ON prism_heartbeats(cell_id);
CREATE INDEX IF NOT EXISTS idx_prism_heartbeats_received ON prism_heartbeats(received_at DESC);

-- Retention: index on received_at for time-based cleanup queries
CREATE INDEX IF NOT EXISTS idx_prism_heartbeats_retention
    ON prism_heartbeats(received_at);

CREATE TABLE IF NOT EXISTS prism_execution_results (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id    TEXT NOT NULL,
    cell_id         TEXT NOT NULL,
    status          TEXT NOT NULL,
    error           TEXT,
    result          JSONB,
    duration_ms     INTEGER,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prism_exec_results_cell ON prism_execution_results(cell_id);
CREATE INDEX IF NOT EXISTS idx_prism_exec_results_execution ON prism_execution_results(execution_id);
CREATE INDEX IF NOT EXISTS idx_prism_exec_results_received ON prism_execution_results(received_at DESC);

CREATE TABLE IF NOT EXISTS prism_capabilities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cell_id         TEXT NOT NULL,
    capability      TEXT NOT NULL,
    trust_score     DOUBLE PRECISION NOT NULL DEFAULT 0,
    metadata        JSONB NOT NULL DEFAULT '{}',
    announced_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(cell_id, capability)
);

CREATE INDEX IF NOT EXISTS idx_prism_capabilities_cell ON prism_capabilities(cell_id);
CREATE INDEX IF NOT EXISTS idx_prism_capabilities_name ON prism_capabilities(capability);

CREATE TABLE IF NOT EXISTS prism_runtime_status (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    healthy         BOOLEAN NOT NULL DEFAULT true,
    active_cells    INTEGER NOT NULL DEFAULT 0,
    active_swarms   INTEGER NOT NULL DEFAULT 0,
    metadata        JSONB NOT NULL DEFAULT '{}',
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prism_status_received ON prism_runtime_status(received_at DESC);
