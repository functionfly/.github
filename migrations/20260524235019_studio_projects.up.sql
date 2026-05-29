-- Studio projects, files, and per-user workspace session state

CREATE TABLE IF NOT EXISTS studio_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    environment TEXT NOT NULL DEFAULT '',
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS studio_project_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES studio_projects(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    path TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    language VARCHAR(50) NOT NULL DEFAULT 'plaintext',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_studio_project_file_path UNIQUE (project_id, path)
);

CREATE TABLE IF NOT EXISTS studio_project_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    environment TEXT NOT NULL DEFAULT '',
    active_project_id UUID REFERENCES studio_projects(id) ON DELETE SET NULL,
    active_file_id UUID REFERENCES studio_project_files(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_studio_project_session UNIQUE (tenant_id, user_id, environment)
);

CREATE INDEX IF NOT EXISTS idx_studio_projects_tenant_user_env
    ON studio_projects(tenant_id, user_id, environment, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_studio_project_files_project
    ON studio_project_files(project_id);
CREATE INDEX IF NOT EXISTS idx_studio_project_files_tenant
    ON studio_project_files(tenant_id);

COMMENT ON TABLE studio_projects IS 'FunctionFly Studio code projects per user and environment';
COMMENT ON TABLE studio_project_files IS 'Source files belonging to a Studio project';
COMMENT ON TABLE studio_project_sessions IS 'Active project/file selection for Studio workspace restore';
