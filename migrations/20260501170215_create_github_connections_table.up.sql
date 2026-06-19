-- GitHub Connections: Linked GitHub accounts with encrypted OAuth tokens
CREATE TABLE IF NOT EXISTS github_connections (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    github_user_id      BIGINT NOT NULL,
    github_username     VARCHAR(255) NOT NULL,
    github_avatar_url   TEXT,
    github_profile_url  TEXT,
    encrypted_token     TEXT NOT NULL,
    token_iv            TEXT NOT NULL,
    token_tag           TEXT NOT NULL,
    encrypted_refresh   TEXT,
    refresh_iv          TEXT,
    refresh_tag         TEXT,
    token_scope         VARCHAR(500),
    token_expires_at    TIMESTAMPTZ,
    github_app_install  BOOLEAN DEFAULT FALSE,
    github_install_id   BIGINT,
    status              VARCHAR(50) NOT NULL DEFAULT 'active',
    last_synced_at      TIMESTAMPTZ,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, github_user_id)
);

CREATE INDEX idx_gh_conn_user ON github_connections(user_id);
CREATE INDEX idx_gh_conn_tenant ON github_connections(tenant_id);
CREATE INDEX idx_gh_conn_status ON github_connections(status) WHERE status = 'active';

-- GitHub Repos: Cached repository metadata
CREATE TABLE IF NOT EXISTS github_repos (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id       UUID NOT NULL REFERENCES github_connections(id) ON DELETE CASCADE,
    github_repo_id      BIGINT NOT NULL,
    full_name           VARCHAR(500) NOT NULL,
    name                VARCHAR(255) NOT NULL,
    owner               VARCHAR(255) NOT NULL,
    description         TEXT,
    default_branch      VARCHAR(255) NOT NULL DEFAULT 'main',
    language            VARCHAR(100),
    languages           JSONB DEFAULT '{}',
    is_private          BOOLEAN NOT NULL DEFAULT FALSE,
    is_fork             BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived         BOOLEAN NOT NULL DEFAULT FALSE,
    topics              JSONB DEFAULT '[]',
    stars_count         INT DEFAULT 0,
    forks_count         INT DEFAULT 0,
    size_kb             INT DEFAULT 0,
    pushed_at           TIMESTAMPTZ,
    html_url            TEXT NOT NULL,
    clone_url           TEXT NOT NULL,
    ssh_url             TEXT NOT NULL,
    detected_functions  JSONB DEFAULT '[]',
    detected_runtime    VARCHAR(50),
    has_functionfly_json BOOLEAN DEFAULT FALSE,
    import_status       VARCHAR(50) DEFAULT 'not_imported',
    metadata            JSONB DEFAULT '{}',
    last_scanned_at     TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(connection_id, github_repo_id)
);

CREATE INDEX idx_gh_repo_conn ON github_repos(connection_id);
CREATE INDEX idx_gh_repo_status ON github_repos(import_status);
CREATE INDEX idx_gh_repo_full_name ON github_repos(full_name);

-- GitHub Imports: Import jobs and history
CREATE TABLE IF NOT EXISTS github_imports (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    connection_id       UUID NOT NULL REFERENCES github_connections(id),
    repo_id             UUID NOT NULL REFERENCES github_repos(id),
    source_branch       VARCHAR(255) NOT NULL DEFAULT 'main',
    source_path         TEXT,
    function_name       VARCHAR(255) NOT NULL,
    function_id         UUID REFERENCES registry_functions(id),
    function_version_id UUID REFERENCES registry_function_versions(id),
    visibility          VARCHAR(50) NOT NULL DEFAULT 'private',
    runtime_override    VARCHAR(50),
    manifest_overrides  JSONB DEFAULT '{}',
    auto_sync_enabled   BOOLEAN DEFAULT FALSE,
    sync_branches       JSONB DEFAULT '["main"]',
    environment_mappings JSONB DEFAULT '{}',
    status              VARCHAR(50) NOT NULL DEFAULT 'pending',
    progress            INT DEFAULT 0,
    error_message       TEXT,
    error_details       JSONB,
    content_hash        VARCHAR(64),
    commit_sha          VARCHAR(40),
    files_imported      INT DEFAULT 0,
    total_size_bytes    BIGINT DEFAULT 0,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ
);

CREATE INDEX idx_gh_import_user ON github_imports(user_id);
CREATE INDEX idx_gh_import_tenant ON github_imports(tenant_id);
CREATE INDEX idx_gh_import_repo ON github_imports(repo_id);
CREATE INDEX idx_gh_import_status ON github_imports(status);
CREATE INDEX idx_gh_import_function ON github_imports(function_id) WHERE function_id IS NOT NULL;

-- GitHub Webhooks: Registered webhooks for auto-sync
CREATE TABLE IF NOT EXISTS github_webhooks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id       UUID NOT NULL REFERENCES github_connections(id) ON DELETE CASCADE,
    repo_id             UUID NOT NULL REFERENCES github_repos(id) ON DELETE CASCADE,
    github_webhook_id   BIGINT,
    webhook_secret      VARCHAR(255) NOT NULL,
    events              JSONB NOT NULL DEFAULT '["push"]',
    is_active           BOOLEAN DEFAULT TRUE,
    last_delivery_at    TIMESTAMPTZ,
    last_event_type     VARCHAR(100),
    delivery_count      INT DEFAULT 0,
    error_count         INT DEFAULT 0,
    last_error          TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(repo_id, events)
);

-- GitHub Sync Logs: Push-to-deploy history
CREATE TABLE IF NOT EXISTS github_sync_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    import_id           UUID NOT NULL REFERENCES github_imports(id) ON DELETE CASCADE,
    function_id         UUID REFERENCES registry_functions(id),
    trigger_type        VARCHAR(50) NOT NULL,
    trigger_branch      VARCHAR(255),
    trigger_commit_sha  VARCHAR(40),
    trigger_pr_number   INT,
    status              VARCHAR(50) NOT NULL DEFAULT 'pending',
    version_published   VARCHAR(50),
    status_check_url    TEXT,
    duration_ms         INT,
    error_message       TEXT,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ
);

CREATE INDEX idx_gh_sync_import ON github_sync_logs(import_id);
CREATE INDEX idx_gh_sync_function ON github_sync_logs(function_id);
CREATE INDEX idx_gh_sync_status ON github_sync_logs(status);
CREATE INDEX idx_gh_sync_created ON github_sync_logs(created_at DESC);

-- GitHub Import Templates: Reusable import configurations
CREATE TABLE IF NOT EXISTS github_import_templates (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id),
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    config              JSONB NOT NULL,
    detection_rules     JSONB DEFAULT '{}',
    is_default          BOOLEAN DEFAULT FALSE,
    usage_count         INT DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gh_template_tenant ON github_import_templates(tenant_id);
