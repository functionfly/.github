# GitHub Repository Import — Architecture Plan

**Feature:** GitHub Repo-to-Function Import Pipeline
**Status:** Architecture Draft
**Date:** 2026-05-01

---

## 1. Executive Summary

Allow users to link their GitHub account (with extended OAuth scopes), browse their repositories, select repos to import, and auto-create FunctionFly registry functions from the code — with optional bidirectional sync (push-to-deploy), branch-based environments, PR preview deployments, and AI-powered manifest generation.

### Key Innovations Beyond Basic Import

| Innovation | Description |
|-----------|-------------|
| **Smart Runtime Detection** | Auto-detect language/runtime from `package.json`, `go.mod`, `Cargo.toml`, `requirements.txt`, `pyproject.toml`, etc. |
| **Multi-Function Monorepo Support** | Detect multiple functions in a single repo (`/functions/*`, `/packages/*`, `serverless.yml`) |
| **AI Manifest Generation** | Use FlyMind (ai-service on :8081) to analyze code and generate optimal `functionfly.jsonc` manifests |
| **Push-to-Deploy Sync** | GitHub webhooks trigger automatic redeployment on push to tracked branches |
| **PR Preview Deployments** | Each PR gets a unique preview function URL for testing before merge |
| **Branch Environment Mapping** | `main` → production, `staging` → staging, `dev` → development function versions |
| **GitHub Status Checks** | Report verification/trust status back to GitHub commit statuses |
| **Import Templates** | Save and reuse import configurations across similar repos |
| **Dependency-Aware Import** | Analyze and bundle dependencies (npm install, pip install, go mod tidy) during import |
| **Collaborative Import Approval** | Team workflow: propose import → review → approve → deploy |

---

## 2. System Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                         Dashboard (React SPA)                        │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────────────┐ │
│  │ GitHub Link  │  │ Repo Browser │  │ Import Configuration Panel │ │
│  │  (OAuth)     │  │ (selectable) │  │ (per-repo settings)        │ │
│  └──────┬───────┘  └──────┬───────┘  └──────────────┬──────────────┘ │
└─────────┼─────────────────┼─────────────────────────┼────────────────┘
          │                 │                         │
          ▼                 ▼                         ▼
┌──────────────────────────────────────────────────────────────────────┐
│                      Orchestrator API (Go)                           │
│                                                                      │
│  ┌─────────────────────┐  ┌────────────────────┐                    │
│  │ GitHub Connect       │  │ Import Pipeline    │                    │
│  │ Handler              │  │ Handler            │                    │
│  │ - OAuth flow (ext)   │  │ - List repos       │                    │
│  │ - Token management   │  │ - Scan repo        │                    │
│  │ - Account linking    │  │ - Configure import │                    │
│  └─────────┬───────────┘  │ - Execute import   │                    │
│            │               │ - Sync management  │                    │
│            │               └────────┬───────────┘                    │
│            │                        │                                │
│  ┌─────────▼────────────────────────▼───────────┐                   │
│  │          GitHub Integration Service           │                   │
│  │  - Token refresh / vault storage              │                   │
│  │  - GitHub API client (rate-limited)           │                   │
│  │  - Repo scanning & function detection         │                   │
│  │  - Webhook management                         │                   │
│  │  - Status check reporting                     │                   │
│  └─────────┬────────────────────────┬───────────┘                   │
│            │                        │                                │
│  ┌─────────▼──────────┐  ┌─────────▼──────────┐                    │
│  │ GitHub Token Vault  │  │ Import Worker Pool  │                    │
│  │ (AES-256-GCM)      │  │ (buffered channel)  │                    │
│  └────────────────────┘  └─────────┬──────────┘                    │
│                                     │                                │
│  ┌──────────────────────────────────▼──────────┐                    │
│  │         Function Registry Publisher          │                    │
│  │  - Creates registry_functions + versions     │                    │
│  │  - Sets visibility (public/private)          │                    │
│  │  - Charges platform fees                     │                    │
│  │  - Triggers verification pipeline            │                    │
│  └─────────────────────────────────────────────┘                    │
└──────────────────────────────────────────────────────────────────────┘
          │                    │                    │
          ▼                    ▼                    ▼
┌──────────────┐  ┌──────────────────┐  ┌────────────────────┐
│  PostgreSQL   │  │  GitHub API      │  │  FlyMind AI        │
│  - gh_*       │  │  - Repos         │  │  - Manifest gen    │
│    tables     │  │  - Contents      │  │  - Runtime detect  │
│               │  │  - Webhooks      │  │  - Dep analysis    │
│               │  │  - Commit status │  │                    │
└──────────────┘  └──────────────────┘  └────────────────────┘
```

---

## 3. Database Schema

### 3.1 `github_connections` — Linked GitHub Accounts

```sql
CREATE TABLE github_connections (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    github_user_id      BIGINT NOT NULL,              -- GitHub's numeric user ID
    github_username     VARCHAR(255) NOT NULL,         -- e.g. "octocat"
    github_avatar_url   TEXT,
    github_profile_url  TEXT,
    encrypted_token     TEXT NOT NULL,                 -- AES-256-GCM encrypted access token
    token_iv            TEXT NOT NULL,                 -- Initialization vector
    token_tag           TEXT NOT NULL,                 -- GCM auth tag
    encrypted_refresh   TEXT,                          -- encrypted refresh token (if applicable)
    refresh_iv          TEXT,
    refresh_tag         TEXT,
    token_scope         VARCHAR(500),                  -- granted scopes: "repo,read:user,user:email"
    token_expires_at    TIMESTAMPTZ,
    github_app_install  BOOLEAN DEFAULT FALSE,         -- true if also installed GitHub App
    github_install_id   BIGINT,                        -- GitHub App installation ID
    status              VARCHAR(50) NOT NULL DEFAULT 'active',  -- active, expired, revoked, error
    last_synced_at      TIMESTAMPTZ,
    metadata            JSONB DEFAULT '{}',            -- rate limit info, plan, etc.
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id, github_user_id)
);

CREATE INDEX idx_gh_conn_user ON github_connections(user_id);
CREATE INDEX idx_gh_conn_tenant ON github_connections(tenant_id);
CREATE INDEX idx_gh_conn_status ON github_connections(status) WHERE status = 'active';
```

### 3.2 `github_repos` — Cached Repository Metadata

```sql
CREATE TABLE github_repos (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id       UUID NOT NULL REFERENCES github_connections(id) ON DELETE CASCADE,
    github_repo_id      BIGINT NOT NULL,              -- GitHub's numeric repo ID
    full_name           VARCHAR(500) NOT NULL,         -- e.g. "octocat/hello-world"
    name                VARCHAR(255) NOT NULL,         -- e.g. "hello-world"
    owner               VARCHAR(255) NOT NULL,         -- e.g. "octocat"
    description         TEXT,
    default_branch      VARCHAR(255) NOT NULL DEFAULT 'main',
    language            VARCHAR(100),                  -- primary language
    languages           JSONB DEFAULT '{}',            -- {"Go": 75.5, "Shell": 24.5}
    is_private          BOOLEAN NOT NULL DEFAULT FALSE,
    is_fork             BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived         BOOLEAN NOT NULL DEFAULT FALSE,
    topics              JSONB DEFAULT '[]',            -- ["serverless", "api"]
    stars_count         INT DEFAULT 0,
    forks_count         INT DEFAULT 0,
    size_kb             INT DEFAULT 0,
    pushed_at           TIMESTAMPTZ,                   -- last push timestamp
    html_url            TEXT NOT NULL,
    clone_url           TEXT NOT NULL,
    ssh_url             TEXT NOT NULL,
    detected_functions  JSONB DEFAULT '[]',            -- auto-detected function paths
    detected_runtime    VARCHAR(50),                   -- auto-detected primary runtime
    has_functionfly_json BOOLEAN DEFAULT FALSE,        -- repo already has functionfly.jsonc
    import_status       VARCHAR(50) DEFAULT 'not_imported', -- not_imported, importing, imported, error
    metadata            JSONB DEFAULT '{}',
    last_scanned_at     TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(connection_id, github_repo_id)
);

CREATE INDEX idx_gh_repo_conn ON github_repos(connection_id);
CREATE INDEX idx_gh_repo_status ON github_repos(import_status);
CREATE INDEX idx_gh_repo_full_name ON github_repos(full_name);
```

### 3.3 `github_imports` — Import Jobs & History

```sql
CREATE TABLE github_imports (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id),
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    connection_id       UUID NOT NULL REFERENCES github_connections(id),
    repo_id             UUID NOT NULL REFERENCES github_repos(id),

    -- What was imported
    source_branch       VARCHAR(255) NOT NULL DEFAULT 'main',
    source_path         TEXT,                          -- subdirectory for monorepo functions
    function_name       VARCHAR(255) NOT NULL,         -- resulting function name
    function_id         UUID REFERENCES registry_functions(id), -- created function
    function_version_id UUID REFERENCES registry_function_versions(id),

    -- Import configuration
    visibility          VARCHAR(50) NOT NULL DEFAULT 'private',  -- public, private, unlisted
    runtime_override    VARCHAR(50),                   -- manual runtime override
    manifest_overrides  JSONB DEFAULT '{}',            -- manual manifest field overrides
    auto_sync_enabled   BOOLEAN DEFAULT FALSE,         -- push-to-deploy
    sync_branches       JSONB DEFAULT '["main"]',      -- branches that trigger sync
    environment_mappings JSONB DEFAULT '{}',           -- {"main": "production", "staging": "staging"}

    -- Status tracking
    status              VARCHAR(50) NOT NULL DEFAULT 'pending',
        -- pending, scanning, configuring, building, publishing, completed, failed, cancelled
    progress            INT DEFAULT 0,                 -- 0-100
    error_message       TEXT,
    error_details       JSONB,

    -- Results
    content_hash        VARCHAR(64),                   -- SHA-256 of imported content
    commit_sha          VARCHAR(40),                   -- GitHub commit SHA imported
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
```

### 3.4 `github_webhooks` — Registered Webhooks

```sql
CREATE TABLE github_webhooks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id       UUID NOT NULL REFERENCES github_connections(id) ON DELETE CASCADE,
    repo_id             UUID NOT NULL REFERENCES github_repos(id) ON DELETE CASCADE,
    github_webhook_id   BIGINT,                        -- GitHub's webhook ID
    webhook_secret      VARCHAR(255) NOT NULL,         -- HMAC signing secret (encrypted)
    events              JSONB NOT NULL DEFAULT '["push"]',  -- subscribed events
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
```

### 3.5 `github_sync_logs` — Sync/Deploy History

```sql
CREATE TABLE github_sync_logs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    import_id           UUID NOT NULL REFERENCES github_imports(id) ON DELETE CASCADE,
    function_id         UUID REFERENCES registry_functions(id),
    trigger_type        VARCHAR(50) NOT NULL,          -- push, pr_open, pr_sync, pr_close, manual
    trigger_branch      VARCHAR(255),
    trigger_commit_sha  VARCHAR(40),
    trigger_pr_number   INT,
    status              VARCHAR(50) NOT NULL DEFAULT 'pending',
        -- pending, building, deploying, completed, failed, skipped
    version_published   VARCHAR(50),                   -- new version published (e.g. "1.2.3+gh.abc1234")
    status_check_url    TEXT,                           -- GitHub status check URL
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
```

### 3.6 `github_import_templates` — Reusable Import Configs

```sql
CREATE TABLE github_import_templates (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id),
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    -- Template config (reusable across repos)
    config              JSONB NOT NULL,                -- full import configuration
        -- { visibility, runtime, manifest_overrides, sync_branches, env_mappings, ... }
    detection_rules     JSONB DEFAULT '{}',            -- custom rules for function detection
        -- { entry_points: ["src/index.ts"], build_command: "npm run build", ... }
    is_default          BOOLEAN DEFAULT FALSE,
    usage_count         INT DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gh_template_tenant ON github_import_templates(tenant_id);
```

---

## 4. API Endpoints

### 4.1 GitHub Connection Management

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/github/connect` | Required | Initiate GitHub OAuth with extended scopes (`repo`, `read:user`, `user:email`) |
| `GET` | `/v1/github/callback` | — | OAuth callback (exchanges code, stores encrypted token) |
| `GET` | `/v1/github/connection` | Required | Get current GitHub connection status |
| `DELETE` | `/v1/github/connection` | Required | Disconnect GitHub account (revokes webhooks, preserves imports) |
| `POST` | `/v1/github/connection/refresh` | Required | Force token refresh |
| `GET` | `/v1/github/connection/scopes` | Required | Check which scopes are granted |

### 4.2 Repository Browsing & Scanning

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/github/repos` | Required | List user's GitHub repos (paginated, filterable by language/visibility) |
| `POST` | `/v1/github/repos/refresh` | Required | Force re-fetch repo list from GitHub |
| `GET` | `/v1/github/repos/{repoId}` | Required | Get detailed repo info + detected functions |
| `POST` | `/v1/github/repos/{repoId}/scan` | Required | Deep-scan repo for functions (AI-powered) |
| `GET` | `/v1/github/repos/{repoId}/branches` | Required | List branches for a repo |
| `GET` | `/v1/github/repos/{repoId}/tree` | Required | Browse repo file tree (for subdirectory selection) |

### 4.3 Import Operations

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/v1/github/import` | Required | Start a new import job |
| `POST` | `/v1/github/import/bulk` | Required | Bulk import multiple repos at once |
| `GET` | `/v1/github/imports` | Required | List all imports (paginated, filterable) |
| `GET` | `/v1/github/imports/{importId}` | Required | Get import details + progress |
| `POST` | `/v1/github/imports/{importId}/cancel` | Required | Cancel an in-progress import |
| `DELETE` | `/v1/github/imports/{importId}` | Required | Delete import record (keeps function) |
| `POST` | `/v1/github/imports/{importId}/retry` | Required | Retry a failed import |
| `POST` | `/v1/github/imports/{importId}/resync` | Required | Force re-import from latest commit |

### 4.4 Sync Management

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `PUT` | `/v1/github/imports/{importId}/sync` | Required | Update sync settings (branches, auto-sync toggle) |
| `GET` | `/v1/github/imports/{importId}/sync-logs` | Required | Get sync/deploy history |
| `POST` | `/v1/github/webhook` | — (HMAC verified) | GitHub webhook receiver (push, PR events) |

### 4.5 Import Templates

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/v1/github/templates` | Required | Create import template |
| `GET` | `/v1/github/templates` | Required | List templates |
| `PUT` | `/v1/github/templates/{id}` | Required | Update template |
| `DELETE` | `/v1/github/templates/{id}` | Required | Delete template |

### 4.6 Admin Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/admin/github/connections` | Admin | List all GitHub connections |
| `GET` | `/v1/admin/github/imports` | Admin | List all imports across tenants |
| `POST` | `/v1/admin/github/imports/{id}/retry` | Admin | Admin retry of failed import |
| `GET` | `/v1/admin/github/stats` | Admin | Import statistics & health |

---

## 5. Service Layer Architecture

### 5.1 New Packages

```
internal/
├── services/
│   └── github/
│       ├── client.go              # GitHub API client (rate-limited, retry, pagination)
│       ├── client_test.go
│       ├── connection.go          # Token management, OAuth extended flow
│       ├── connection_test.go
│       ├── repos.go               # Repo listing, metadata caching
│       ├── repos_test.go
│       ├── scanner.go             # Function detection engine
│       ├── scanner_test.go
│       ├── importer.go            # Import pipeline orchestration
│       ├── importer_test.go
│       ├── syncer.go              # Push-to-deploy sync engine
│       ├── syncer_test.go
│       ├── webhook_manager.go     # GitHub webhook CRUD
│       ├── webhook_manager_test.go
│       ├── status_checks.go       # Commit status reporting
│       ├── status_checks_test.go
│       ├── manifest_generator.go  # AI-powered manifest generation
│       ├── manifest_generator_test.go
│       ├── token_vault.go         # Encrypted token storage
│       └── types.go               # Shared types
├── api/
│   └── handlers/
│       └── github/
│       ├── handler.go             # Main handler struct
│       ├── connection.go          # Connection endpoints
│       ├── repos.go               # Repo browsing endpoints
│       ├── import.go              # Import operation endpoints
│       ├── sync.go                # Sync management endpoints
│       ├── webhook.go             # Webhook receiver
│       ├── templates.go           # Import template endpoints
│       └── types.go               # Request/response types
└── storage/
    └── github/
    ├── repository.go              # Main repository interface
    ├── connection_repo.go         # github_connections CRUD
    ├── repo_repo.go               # github_repos CRUD
    ├── import_repo.go             # github_imports CRUD
    ├── webhook_repo.go            # github_webhooks CRUD
    ├── sync_log_repo.go           # github_sync_logs CRUD
    ├── template_repo.go           # github_import_templates CRUD
    └── types.go                   # DB model types
```

### 5.2 GitHub API Client (`internal/services/github/client.go`)

```go
type Client struct {
    httpClient    *http.Client
    baseURL       string
    rateLimiter   *RateLimiter
    logger        *logrus.Logger
    userAgent     string
}

// Rate limiter respects GitHub's rate limit headers
type RateLimiter struct {
    remaining int
    resetAt   time.Time
    mu        sync.Mutex
}

// Key methods
func (c *Client) GetAuthenticatedUser(ctx context.Context) (*GitHubUser, error)
func (c *Client) ListRepos(ctx context.Context, opts ListReposOpts) ([]*GitHubRepo, error)
func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*GitHubRepo, error)
func (c *Client) GetRepoContent(ctx context.Context, owner, repo, path, ref string) ([]*GitHubContent, error)
func (c *Client) GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, error)
func (c *Client) GetLanguages(ctx context.Context, owner, repo string) (map[string]float64, error)
func (c *Client) ListBranches(ctx context.Context, owner, repo string) ([]*GitHubBranch, error)
func (c *Client) GetTree(ctx context.Context, owner, repo, sha string, recursive bool) ([]*GitHubTreeEntry, error)
func (c *Client) CreateWebhook(ctx context.Context, owner, repo string, req *CreateWebhookRequest) (*GitHubWebhook, error)
func (c *Client) DeleteWebhook(ctx context.Context, owner, repo string, webhookID int64) error
func (c *Client) CreateCommitStatus(ctx context.Context, owner, repo, sha string, status *CommitStatus) error
func (c *Client) GetCompareDiff(ctx context.Context, owner, repo, base, head string) (*CompareResult, error)
```

### 5.3 Function Detection Engine (`internal/services/github/scanner.go`)

The scanner uses a **multi-strategy detection pipeline**:

```
┌─────────────────────────────────────────────────────┐
│              Function Detection Pipeline             │
│                                                      │
│  1. Explicit Config Detection                        │
│     └─ functionfly.jsonc / functionfly.json          │
│                                                      │
│  2. Serverless Framework Detection                   │
│     └─ serverless.yml / serverless.yaml              │
│                                                      │
│  3. Framework Pattern Detection                      │
│     ├─ AWS Lambda: src/*handler*, src/*lambda*        │
│     ├─ Cloudflare Workers: wrangler.toml             │
│     ├─ Vercel Functions: api/ directory              │
│     ├─ Netlify Functions: netlify/functions/          │
│     └─ Express/Fastify/Hono: src/index.ts|js         │
│                                                      │
│  4. Entry Point Detection                            │
│     ├─ Node: index.ts/js, main.ts/js, handler.ts/js │
│     ├─ Python: main.py, handler.py, app.py, lambda_ │
│     ├─ Go: main.go, cmd/*/main.go                    │
│     ├─ Rust: src/main.rs, src/lib.rs                 │
│     └─ Generic: README.md heuristic analysis         │
│                                                      │
│  5. AI-Powered Analysis (fallback)                   │
│     └─ FlyMind analyzes repo structure + README      │
│        to suggest function entry points              │
└─────────────────────────────────────────────────────┘
```

```go
type DetectionResult struct {
    Functions      []DetectedFunction `json:"functions"`
    PrimaryRuntime string            `json:"primary_runtime"`
    Confidence     float64           `json:"confidence"`       // 0.0 - 1.0
    Strategy       string            `json:"strategy"`         // which detector found it
    Warnings       []string          `json:"warnings,omitempty"`
}

type DetectedFunction struct {
    Name         string            `json:"name"`
    EntryPoint   string            `json:"entry_point"`       // e.g. "src/index.ts"
    Runtime      string            `json:"runtime"`
    SubDirectory string            `json:"sub_directory,omitempty"` // for monorepos
    Manifest     *FunctionManifest `json:"manifest,omitempty"` // pre-generated manifest
    Dependencies *DependencyInfo   `json:"dependencies,omitempty"`
    Confidence   float64           `json:"confidence"`
}

type Detector interface {
    Name() string
    Detect(ctx context.Context, repo *GitHubRepo, tree []*GitHubTreeEntry) (*DetectionResult, error)
    Priority() int // lower = checked first
}
```

### 5.4 Import Pipeline (`internal/services/github/importer.go`)

The import pipeline is an **async multi-stage processor** using the existing buffered channel + worker pool pattern:

```
┌──────────────────────────────────────────────────────────────┐
│                    Import Pipeline Stages                     │
│                                                              │
│  Stage 1: SCAN (2-5s)                                        │
│  ├─ Fetch repo tree from GitHub API                          │
│  ├─ Run detection pipeline (all 5 strategies)                │
│  ├─ Cache results in github_repos.detected_functions         │
│  └─ Return detection results to user for confirmation        │
│                                                              │
│  Stage 2: CONFIGURE (user-driven, async)                     │
│  ├─ User selects which functions to import                   │
│  ├─ User sets visibility per-function                        │
│  ├─ User optionally overrides runtime/manifest               │
│  ├─ User configures sync settings                            │
│  └─ User optionally applies import template                  │
│                                                              │
│  Stage 3: FETCH (5-30s)                                      │
│  ├─ Download source files from GitHub                        │
│  ├─ Resolve and bundle dependencies                          │
│  │   ├─ Node: npm ci --production && bundle                   │
│  │   ├─ Python: pip install -r requirements.txt              │
│  │   ├─ Go: go mod download                                  │
│  │   └─ Rust: cargo build --release                          │
│  ├─ Generate functionfly.jsonc if not present                │
│  │   └─ Call FlyMind AI for manifest generation              │
│  ├─ Calculate content hash (SHA-256)                         │
│  └─ Store fetched artifacts temporarily                      │
│                                                              │
│  Stage 4: BUILD (10-120s)                                    │
│  ├─ Compile to WASM (if applicable)                          │
│  ├─ Bundle source + dependencies                             │
│  ├─ Run security scan (YARA/ClamAV)                          │
│  └─ Validate manifest                                        │
│                                                              │
│  Stage 5: PUBLISH (2-10s)                                    │
│  ├─ Create registry_function record                          │
│  ├─ Create registry_function_version record                  │
│  ├─ Set visibility (public/private/unlisted)                 │
│  ├─ Charge platform fee (if applicable)                      │
│  ├─ Trigger verification pipeline                            │
│  ├─ Create github_imports record (completed)                 │
│  ├─ Register GitHub webhook (if auto_sync enabled)           │
│  └─ Report GitHub commit status                              │
└──────────────────────────────────────────────────────────────┘
```

```go
type ImportPipeline struct {
    githubClient  *Client
    scanner       *Scanner
    aiService     *FlyMindClient
    publisher     *RegistryPublisher
    webhookMgr    *WebhookManager
    statusChecker *StatusCheckReporter
    repo          storage.GitHubRepository
    logger        *logrus.Logger

    workers    int
    importChan chan *ImportJob
    done       chan struct{}
    wg         sync.WaitGroup
}

type ImportJob struct {
    ID           uuid.UUID
    UserID       uuid.UUID
    TenantID     uuid.UUID
    ConnectionID uuid.UUID
    RepoID       uuid.UUID
    Config       ImportConfig
    Progress     chan ImportProgress
}

type ImportConfig struct {
    Branch          string                       `json:"branch"`
    SubPath         string                       `json:"sub_path,omitempty"`
    Functions       []FunctionImportConfig       `json:"functions"`      // which functions to import
    Visibility      string                       `json:"visibility"`     // default for all
    AutoSync        bool                         `json:"auto_sync"`
    SyncBranches    []string                     `json:"sync_branches"`
    EnvMappings     map[string]string            `json:"env_mappings"`   // branch → environment
    TemplateID      *uuid.UUID                   `json:"template_id,omitempty"`
}

type FunctionImportConfig struct {
    DetectedName     string            `json:"detected_name"`
    CustomName       string            `json:"custom_name,omitempty"`
    EntryPoint       string            `json:"entry_point"`
    SubDirectory     string            `json:"sub_directory,omitempty"`
    Runtime          string            `json:"runtime"`
    Visibility       string            `json:"visibility"`
    ManifestOverride map[string]interface{} `json:"manifest_override,omitempty"`
}
```

### 5.5 Push-to-Deploy Sync Engine (`internal/services/github/syncer.go`)

```
┌──────────────────────────────────────────────────────────────┐
│                    Webhook → Sync Flow                        │
│                                                              │
│  GitHub Push Event                                           │
│       │                                                      │
│       ▼                                                      │
│  POST /v1/github/webhook                                     │
│       │                                                      │
│       ├─ Verify HMAC signature                               │
│       ├─ Parse event payload                                 │
│       ├─ Match repo → github_imports (auto_sync=true)        │
│       │                                                      │
│       ▼                                                      │
│  For each matching import:                                   │
│       ├─ Check if pushed branch is in sync_branches          │
│       ├─ Get diff (compare base_sha..new_sha)                │
│       │   ├─ If no relevant changes → skip (log it)          │
│       │   └─ If function code changed → proceed              │
│       ├─ Create github_sync_log (pending)                    │
│       ├─ Report GitHub commit status: "pending"              │
│       │                                                      │
│       ▼                                                      │
│  Import Pipeline (stages 3-5 only, incremental)              │
│       ├─ Fetch only changed files (sparse checkout)          │
│       ├─ Re-build affected functions                          │
│       ├─ Publish new version (auto-increment: semver+sha)    │
│       │   └─ e.g. "1.2.3+gh.abc1234"                        │
│       ├─ Report GitHub commit status: "success" / "failure"  │
│       └─ Update sync log                                     │
│                                                              │
│  PR Events (opened/synchronized):                            │
│       ├─ Create preview version (unlisted, TTL 7 days)       │
│       ├─ Generate preview URL                                │
│       ├─ Post PR comment with preview link                   │
│       └─ Report commit status with preview URL               │
│                                                              │
│  PR Events (closed/merged):                                  │
│       ├─ If merged → trigger full sync on target branch      │
│       └─ If closed → cleanup preview version (soft delete)   │
└──────────────────────────────────────────────────────────────┘
```

---

## 6. GitHub OAuth Extended Scopes

The current OAuth flow uses `read:user` + `user:email`. For repo import, we need an **extended scope flow** that's separate from login:

### 6.1 Two-Tier OAuth Approach

| Tier | Scopes | Purpose | When |
|------|--------|---------|------|
| **Login** (existing) | `read:user`, `user:email` | Authentication only | Login/signup |
| **Connect** (new) | `repo`, `read:user`, `user:email` | Full repo access | Linking for import |

### 6.2 Why `repo` Scope and Not GitHub App

| Approach | Pros | Cons |
|----------|------|------|
| **OAuth `repo` scope** | Simple, no app install, works for all repos immediately | Token expires, broad access, no fine-grained permissions |
| **GitHub App (fine-grained)** | Per-repo access, installation tokens, no user token needed | Complex setup, requires install per-org, UI friction |
| **Hybrid (recommended)** | OAuth for user auth + GitHub App for production sync | Most complex to implement |

**Recommendation:** Start with OAuth `repo` scope for MVP. Add GitHub App support later for enterprise customers who need per-repo fine-grained access.

### 6.3 Token Lifecycle

```
┌──────────────────────────────────────────────────────────┐
│                  Token Lifecycle                          │
│                                                          │
│  1. User clicks "Connect GitHub"                         │
│     └─ OAuth flow with scopes: repo,read:user,user:email │
│                                                          │
│  2. Callback receives auth code                          │
│     ├─ Exchange code for access_token + refresh_token    │
│     ├─ Encrypt tokens (AES-256-GCM)                      │
│     ├─ Store in github_connections table                  │
│     └─ Cache repo list (async)                           │
│                                                          │
│  3. Token Refresh (background)                           │
│     ├─ Check token_expires_at hourly                     │
│     ├─ If < 1 hour until expiry → refresh                │
│     ├─ Update encrypted token in DB                      │
│     └─ If refresh fails → mark connection 'expired'      │
│        └─ Notify user to re-authorize                    │
│                                                          │
│  4. Token Revocation                                     │
│     ├─ User disconnects → DELETE /v1/github/connection   │
│     ├─ Revoke webhooks on GitHub                         │
│     ├─ Delete cached repos                               │
│     └─ Preserve import records + created functions       │
└──────────────────────────────────────────────────────────┘
```

---

## 7. Dashboard UI Flow

### 7.1 New Pages

| Route | Component | Description |
|-------|-----------|-------------|
| `/settings/github` | `GitHubSettingsPage` | Connection management, disconnect |
| `/github/repos` | `GitHubReposPage` | Browse repos, trigger scan |
| `/github/repos/:repoId/import` | `ImportConfigPage` | Configure and start import |
| `/github/imports` | `ImportsPage` | List all imports, status, manage sync |
| `/github/imports/:importId` | `ImportDetailPage` | Import details, sync logs, linked function |
| `/github/templates` | `ImportTemplatesPage` | Manage import templates |

### 7.2 Key UI Components

#### Repo Browser Card

```
┌─────────────────────────────────────────────────────────┐
│  📦 octocat/hello-world                    ⭐ 1.2k     │
│  A friendly greeting function                           │
│  ┌────────┐ ┌──────────┐ ┌─────────┐                   │
│  │ Go 85% │ │ Shell 15%│ │ Private │                   │
│  └────────┘ └──────────┘ └─────────┘                   │
│                                                         │
│  🔍 Detected Functions:                                 │
│  ┌─────────────────────────────────────────────────┐   │
│  │ ☑ greet-handler     src/handler.go    [Go]     │   │
│  │ ☑ health-check      src/health.go     [Go]     │   │
│  │ □ test-utils        src/test/utils.go [Go]     │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
│  Default Branch: main    Last Push: 2 hours ago         │
│                                                         │
│  [ Import Selected (2) ]  [ Import All ]  [ Configure ] │
└─────────────────────────────────────────────────────────┘
```

#### Import Progress Stepper

```
┌──────────────────────────────────────────────────────────┐
│  Importing: octocat/hello-world → hello-world-handler    │
│                                                          │
│  ● Scanning          ✓ Complete                          │
│  ● Configuring       ✓ Complete                          │
│  ● Fetching          ✓ Complete                          │
│  ● Building          ◐ In Progress... (67%)              │
│  ○ Publishing        Pending                              │
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Building WASM bundle...                          │   │
│  │  ████████████████████░░░░░░  67%                  │   │
│  │  Compiling handler.go → wasm (4.2 MB)             │   │
│  └──────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────┘
```

---

## 8. Route Registration

Add to `internal/api/routes.go` or a new `routes_github.go`:

```go
func (s *Server) registerGitHubRoutes(api *mux.Router, authMiddleware *middleware.AuthMiddleware) {
    github := api.PathPrefix("/v1/github").Subrouter()

    // Connection management
    github.HandleFunc("/connect", authMiddleware.RequireAuth(s.githubHandler.HandleConnect)).Methods("GET")
    github.HandleFunc("/callback", s.githubHandler.HandleCallback).Methods("GET") // no auth - OAuth callback
    github.HandleFunc("/connection", authMiddleware.RequireAuth(s.githubHandler.HandleGetConnection)).Methods("GET")
    github.HandleFunc("/connection", authMiddleware.RequireAuth(s.githubHandler.HandleDisconnect)).Methods("DELETE")
    github.HandleFunc("/connection/refresh", authMiddleware.RequireAuth(s.githubHandler.HandleRefreshToken)).Methods("POST")

    // Repos
    github.HandleFunc("/repos", authMiddleware.RequireAuth(s.githubHandler.HandleListRepos)).Methods("GET")
    github.HandleFunc("/repos/refresh", authMiddleware.RequireAuth(s.githubHandler.HandleRefreshRepos)).Methods("POST")
    github.HandleFunc("/repos/{repoId}", authMiddleware.RequireAuth(s.githubHandler.HandleGetRepo)).Methods("GET")
    github.HandleFunc("/repos/{repoId}/scan", authMiddleware.RequireAuth(s.githubHandler.HandleScanRepo)).Methods("POST")
    github.HandleFunc("/repos/{repoId}/branches", authMiddleware.RequireAuth(s.githubHandler.HandleListBranches)).Methods("GET")
    github.HandleFunc("/repos/{repoId}/tree", authMiddleware.RequireAuth(s.githubHandler.HandleGetTree)).Methods("GET")

    // Imports
    github.HandleFunc("/import", authMiddleware.RequireAuth(s.githubHandler.HandleImport)).Methods("POST")
    github.HandleFunc("/import/bulk", authMiddleware.RequireAuth(s.githubHandler.HandleBulkImport)).Methods("POST")
    github.HandleFunc("/imports", authMiddleware.RequireAuth(s.githubHandler.HandleListImports)).Methods("GET")
    github.HandleFunc("/imports/{importId}", authMiddleware.RequireAuth(s.githubHandler.HandleGetImport)).Methods("GET")
    github.HandleFunc("/imports/{importId}/cancel", authMiddleware.RequireAuth(s.githubHandler.HandleCancelImport)).Methods("POST")
    github.HandleFunc("/imports/{importId}/retry", authMiddleware.RequireAuth(s.githubHandler.HandleRetryImport)).Methods("POST")
    github.HandleFunc("/imports/{importId}/resync", authMiddleware.RequireAuth(s.githubHandler.HandleResyncImport)).Methods("POST")
    github.HandleFunc("/imports/{importId}/sync", authMiddleware.RequireAuth(s.githubHandler.HandleUpdateSync)).Methods("PUT")
    github.HandleFunc("/imports/{importId}/sync-logs", authMiddleware.RequireAuth(s.githubHandler.HandleGetSyncLogs)).Methods("GET")

    // Templates
    github.HandleFunc("/templates", authMiddleware.RequireAuth(s.githubHandler.HandleListTemplates)).Methods("GET")
    github.HandleFunc("/templates", authMiddleware.RequireAuth(s.githubHandler.HandleCreateTemplate)).Methods("POST")
    github.HandleFunc("/templates/{id}", authMiddleware.RequireAuth(s.githubHandler.HandleUpdateTemplate)).Methods("PUT")
    github.HandleFunc("/templates/{id}", authMiddleware.RequireAuth(s.githubHandler.HandleDeleteTemplate)).Methods("DELETE")

    // Webhook (no auth - HMAC verified)
    github.HandleFunc("/webhook", s.githubHandler.HandleWebhook).Methods("POST")
}
```

---

## 9. Security Considerations

### 9.1 Token Security
- All GitHub tokens encrypted at rest with AES-256-GCM (reuse existing `EncryptionManager`)
- Tokens never logged, never returned to frontend
- Token refresh happens server-side only
- Tokens scoped minimally — `repo` scope for import, could use fine-grained tokens in future

### 9.2 Webhook Security
- GitHub webhook payloads verified with HMAC-SHA256 (reuse existing webhook verification pattern)
- Webhook secrets stored encrypted in `github_webhooks.webhook_secret`
- Replay protection: delivery ID deduplication via Redis (1-hour TTL)
- IP allowlisting for GitHub webhook IPs (optional, configurable)

### 9.3 Import Security
- All imported code goes through YARA/ClamAV malware scanning (reuse existing verification pipeline)
- Dependency resolution happens in isolated containers
- Content hash verification prevents tampering
- Rate limiting: max 10 concurrent imports per tenant
- Quota: imports count toward function quota on plan

### 9.4 Rate Limiting
| Operation | Limit | Window |
|-----------|-------|--------|
| Repo scan | 10 | per minute per user |
| Import start | 5 | per minute per user |
| Webhook processing | 100 | per minute per tenant |
| GitHub API calls | Respects GitHub's limits (5000/hr) | rolling |

---

## 10. Scaling Strategy

### 10.1 Horizontal Scaling

| Component | Strategy |
|-----------|----------|
| **Import Workers** | Buffer channel → distributed queue (RabbitMQ `github.import.queue`) for multi-instance |
| **Webhook Processing** | Idempotent processing with delivery ID → safe for multiple API instances |
| **Token Refresh** | Distributed lock via Redis (`SETNX github:token:refresh:{conn_id}`) |
| **Repo Scanning** | Cache results in `github_repos` with TTL; invalidate on webhook |

### 10.2 Caching Strategy

| Data | Cache Location | TTL | Invalidation |
|------|---------------|-----|--------------|
| Repo list | PostgreSQL `github_repos` | 1 hour | webhook push / manual refresh |
| Repo tree | Redis `github:tree:{owner}:{repo}:{sha}` | 24 hours | new commit SHA |
| File contents | Redis `github:file:{owner}:{repo}:{path}:{sha}` | 24 hours | new commit SHA |
| Detection results | PostgreSQL `github_repos.detected_functions` | 24 hours | manual rescan |
| Rate limit state | Redis `github:ratelimit:{user_id}` | until reset | — |

### 10.3 Performance Targets

| Metric | Target |
|--------|--------|
| Repo listing | < 2s (cached), < 5s (fresh from GitHub) |
| Function detection | < 10s for repos < 1000 files |
| Single function import | < 60s |
| Bulk import (10 repos) | < 5 minutes |
| Webhook → sync start | < 5s |
| Full push-to-deploy cycle | < 2 minutes |

---

## 11. Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Database migrations for all `github_*` tables
- [ ] GitHub API client with rate limiting
- [ ] Extended OAuth flow (connection, not login)
- [ ] Token vault (encrypt/decrypt/refresh)
- [ ] Connection management endpoints
- [ ] Repo listing with caching

### Phase 2: Detection & Import (Week 3-4)
- [ ] Function detection engine (5 strategies)
- [ ] Import pipeline (stages 1-5)
- [ ] Import endpoints + progress tracking (SSE)
- [ ] Basic dashboard UI (repo browser, import config, progress)

### Phase 3: Sync & Webhooks (Week 5-6)
- [ ] GitHub webhook management
- [ ] Push-to-deploy sync engine
- [ ] Commit status reporting
- [ ] Sync management UI

### Phase 4: Innovation Features (Week 7-8)
- [ ] AI-powered manifest generation (FlyMind integration)
- [ ] PR preview deployments
- [ ] Branch environment mapping
- [ ] Import templates
- [ ] Bulk import

### Phase 5: Polish & Scale (Week 9-10)
- [ ] Admin dashboard for GitHub imports
- [ ] Monitoring & alerting
- [ ] Rate limit optimization
- [ ] GitHub App support (enterprise)
- [ ] Documentation & API reference

---

## 12. File Touch List (Existing Files to Modify)

| File | Change |
|------|--------|
| `internal/api/routes.go` | Register GitHub routes (or new `routes_github.go`) |
| `internal/api/middleware/auth.go` | No changes needed — reuse existing patterns |
| `internal/auth/oauth.go` | Add extended scope OAuth flow for `repo` scope |
| `internal/storage/models.go` | Add GitHub model types |
| `migrations/` | 6 new migration files for `github_*` tables |
| `web/dashboard/src/App.tsx` | Add GitHub import routes |
| `web/dashboard/src/api/` | Add `github.ts` API client |
| `web/dashboard/src/hooks/` | Add `useGitHub.ts` query hooks |
| `web/dashboard/src/pages/` | Add GitHub import pages |
| `web/dashboard/src/components/github/` | Add GitHub UI components |

### New Files (estimate: ~40 files)

| Directory | Files | Count |
|-----------|-------|-------|
| `internal/services/github/` | client, connection, repos, scanner, importer, syncer, webhook_manager, status_checks, manifest_generator, token_vault, types | 11 |
| `internal/api/handlers/github/` | handler, connection, repos, import, sync, webhook, templates, types | 8 |
| `internal/storage/github/` | repository, connection_repo, repo_repo, import_repo, webhook_repo, sync_log_repo, template_repo, types | 8 |
| `migrations/` | 6 migration pairs (up/down) | 12 |
| `web/dashboard/src/pages/GitHub*` | 6 page components | 6 |
| `web/dashboard/src/components/github/` | ~10 UI components | 10 |
| `web/dashboard/src/api/github.ts` | API client | 1 |
| `web/dashboard/src/hooks/useGitHub.ts` | Query hooks | 1 |

---

## 13. Testing Strategy

### Unit Tests
- GitHub API client: mock HTTP responses, rate limiter behavior
- Detection engine: fixture repos with known structures
- Import pipeline: mock stages, verify state transitions
- Token vault: encrypt/decrypt round-trip, tamper detection

### Integration Tests
- OAuth flow: mock GitHub OAuth endpoints
- Webhook processing: send mock payloads, verify function updates
- Import end-to-end: mock GitHub API + real PostgreSQL

### E2E Tests (Dashboard)
- Connect GitHub → browse repos → import → verify function created
- Enable auto-sync → push to repo → verify function updated
- Disconnect GitHub → verify functions preserved

---

## 14. Critical Improvements & Suggestions

### 14.1 SSE for Real-Time Import Progress

**Problem:** Import takes 30s–5min. Polling wastes resources and feels laggy.

**Solution:** Add `GET /v1/github/imports/{id}/progress` as an SSE endpoint:
- Backend: Use `http.Flusher` with `text/event-stream` content type
- Frontend: `EventSource` API with `useImportProgress` hook
- Events: `progress` (stage + % + message), `complete` (function_id + version), `error`
- Reconnection: Built into EventSource with automatic retry
- Auth: Pass JWT as query param (EventSource doesn't support headers)

### 14.2 Dry-Run / Preview Mode

**Problem:** Users don't know what will be created or how much it costs before committing.

**Solution:** Add `POST /v1/github/import/preview` endpoint that runs everything except the actual publish:
- Returns: list of functions to be created, visibility, estimated size, estimated cost, conflict detection
- Frontend: `DryRunPreviewDialog` shows the preview table
- User confirms → actual import starts
- Reduces wasted platform fees and user frustration

### 14.3 Conflict Detection & Resolution

**Problem:** Re-importing the same repo or importing a repo with a function name that already exists causes errors.

**Solution:**
- Backend: Before creating a function, check if `author/name` already exists in `registry_functions`
- If conflict detected, return conflict info with resolution options:
  - `overwrite` — replace existing function entirely
  - `rename` — append suffix (e.g., `api-router-v2`)
  - `new_version` — add as new semver version to existing function
  - `skip` — don't import this function
- Frontend: `ConflictResolutionDialog` component for user to choose

### 14.4 GitHub App Support (Enterprise Track)

**Problem:** OAuth `repo` scope gives access to ALL user repos. Enterprise orgs won't accept this.

**Solution:** Design the schema with `github_app_install_id` from day one. Implementation path:
1. **Phase 1 (MVP):** OAuth `repo` scope — works for individual devs
2. **Phase 2 (Growth):** GitHub App with fine-grained permissions — per-repo access, installation tokens
3. **Phase 3 (Enterprise):** GitHub App with org-level installation — team-shared connections

The `github_connections` table already has `github_app_install` and `github_install_id` columns.

### 14.5 `functionfly.jsonc` PR Generation

**Problem:** Repos without a `functionfly.jsonc` manifest rely on AI detection, which may be imperfect.

**Solution:** After import, offer to create a PR back to the source repo with a generated `functionfly.jsonc`:
- Backend: Use GitHub API to create a branch, commit the manifest, open a PR
- Frontend: "Create PR with manifest" button on the import success screen
- Benefits: Future imports are deterministic, drives adoption, creates a link back to FunctionFly

### 14.6 Workspace / Team-Level Connections

**Problem:** Currently connections are per-user. Teams want shared GitHub access.

**Solution:** Add optional `team_id` column to `github_connections`:
- `user_id` is always set (who authorized)
- `team_id` is optional (if team-scoped)
- Team admins can manage team connections
- Import permissions: team members can import using team connections

### 14.7 Import Quota Enforcement

**Problem:** Without limits, abuse is possible (mass importing to inflate registry).

**Solution:** Add plan-based import quotas:
- Free: 5 imports/month, 10 total functions
- Pro: 50 imports/month, 100 total functions
- Enterprise: Unlimited
- Enforce via `RequireFeature` middleware + usage tracking in `tenant_usage` table

### 14.8 Dependency Caching Layer

**Problem:** Every sync re-downloads `node_modules`, `pip` packages, etc. — slow and wasteful.

**Solution:** Cache built dependencies in Redis/R2 keyed by lockfile hash:
- `npm install` → hash `package-lock.json` → check cache → skip if hit
- Same for `requirements.txt` (pip), `go.sum` (go), `Cargo.lock` (cargo)
- Cache TTL: 7 days (invalidate on security advisory)
- Reduces sync time from ~60s to ~10s for unchanged dependencies

### 14.9 Rollback on Partial Failure

**Problem:** Bulk import of 10 functions fails at #7 — functions 1-6 are created, 7-10 are not.

**Solution:** Use the existing saga pattern (`TransactionManager.ExecuteSaga`):
- Each function creation is a saga step with a compensating action (delete the function)
- On failure, compensating actions run in reverse order
- User sees: "6 of 10 functions imported. Rolling back..." or "6 of 10 imported, 4 failed. Keep partial?"

### 14.10 Enhanced New Files List

The UI specification adds these files to the plan (see `docs/GITHUB_IMPORT_UI_SPECIFICATION.md`):

| Category | New Files | Count |
|----------|-----------|-------|
| **API Client** | `api/github.ts` | 1 |
| **Types/Schemas** | `types/github.ts` | 1 |
| **Hooks** | `useGitHubConnection.ts`, `useGitHubRepos.ts`, `useGitHubImport.ts`, `useGitHubSync.ts`, `useGitHubTemplates.ts` | 5 |
| **Store** | `stores/githubStore.ts` | 1 |
| **Components** | 22 components in `components/github/` | 22 |
| **Pages** | `GitHubPage/`, `GitHubRepoImportPage/`, `GitHubSettingsPage/` | 3 pages (7 files) |
| **i18n** | Translation keys in `en.json` + other locales | 1 |
| **Backend** | `internal/services/github/` (11), `internal/api/handlers/github/` (8), `internal/storage/github/` (8) | 27 |
| **Migrations** | 6 tables in 1 migration file | 1 |
| **Total** | | **~66 files** |
