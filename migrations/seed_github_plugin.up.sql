-- Seed data for FunctionFly Studio GitHub Integration plugin
-- Run this after migrations are applied

-- Create a marketplace extension for the GitHub plugin
INSERT INTO marketplace_extensions (
    id,
    creator_id,
    plugin_id,
    name,
    version,
    description,
    category,
    icon_url,
    screenshots,
    manifest,
    manifest_url,
    signature,
    verified,
    status,
    featured,
    install_count,
    rating_average,
    rating_count,
    trust_score,
    sandbox_score,
    security_score,
    runtime_score,
    compatibility,
    tags,
    changelog,
    release_notes,
    published_at
) VALUES (
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    '00000000-0000-0000-0000-000000000001',  -- FunctionFly team ID
    NULL,
    'GitHub Integration',
    '1.0.0',
    'Connect your FunctionFly workflows to GitHub repositories. Sync commits, create issues, manage pull requests, and trigger deployments from GitHub Actions webhooks.',
    'integration',
    'https://cdn.functionfly.com/plugins/github/icon.svg',
    ARRAY['https://cdn.functionfly.com/plugins/github/screenshot1.png', 'https://cdn.functionfly.com/plugins/github/screenshot2.png'],
    '{
        "name": "github-integration",
        "version": "1.0.0",
        "entry_point": "index.js",
        "runtime": "nodejs",
        "permissions": [
            {"type": "webhook", "action": "send", "resource": "github.com"},
            {"type": "api", "action": "read", "resource": "repositories"},
            {"type": "api", "action": "write", "resource": "issues"},
            {"type": "api", "action": "write", "resource": "pull_requests"},
            {"type": "api", "action": "read", "resource": "workflows"},
            {"type": "api", "action": "trigger", "resource": "workflows"}
        ],
        "hooks": {
            "pre_deploy": "hooks/pre-deploy.js",
            "post_deploy": "hooks/post-deploy.js",
            "on_error": "hooks/on-error.js"
        },
        "endpoints": [
            {"path": "/webhook/github", "method": "POST", "handler": "webhookHandler"},
            {"path": "/sync/repos", "method": "GET", "handler": "syncReposHandler"},
            {"path": "/create/issue", "method": "POST", "handler": "createIssueHandler"}
        ],
        "ui": {
            "dashboard_component": "GitHubDashboard",
            "settings_component": "GitHubSettings"
        },
        "config_schema": {
            "github_token": {"type": "string", "required": true, "secret": true},
            "webhook_secret": {"type": "string", "required": true, "secret": true},
            "default_repo": {"type": "string", "required": false},
            "auto_sync": {"type": "boolean", "default": false}
        }
    }'::jsonb,
    'https://cdn.functionfly.com/plugins/github/manifest.json',
    'sig_github_integration_1.0.0',
    true,  -- verified
    'published',
    true,  -- featured
    0,     -- install_count (will increment as users install)
    0.00,  -- rating_average
    0,     -- rating_count
    95.00, -- trust_score
    88.00, -- sandbox_score
    92.00, -- security_score
    85.00, -- runtime_score
    '{
        "platforms": ["nodejs", "bun", "deno"],
        "features": ["sandboxed", "network", "webhooks", "async"]
    }'::jsonb,
    ARRAY['github', 'integration', 'vcs', 'ci/cd', 'automation'],
    '## Changelog

### 1.0.0 (2026-05-16)
- Initial release
- GitHub OAuth integration
- Repository sync
- Issue creation and management
- Pull request status tracking
- GitHub Actions webhook support
- Pre/post deploy hooks',
    '## Features

- **Repository Integration**: Connect multiple GitHub repositories and sync workflow state
- **Issue Management**: Create, update, and close GitHub issues from FunctionFly workflows
- **Pull Request Tracking**: Monitor PR status and merge readiness
- **Webhook Support**: Receive GitHub webhooks for repository events
- **Deploy Hooks**: Trigger deployments from GitHub Actions workflow completions
- **Secret Management**: Secure storage for GitHub tokens using zero-knowledge encryption

## Installation

1. Install from the Plugin Store
2. Authorize with your GitHub account
3. Configure webhook secret
4. Start using GitHub actions in your workflows',
    NOW()
) ON CONFLICT (creator_id, name) DO NOTHING;

-- Create the actual plugin entry for the GitHub plugin
INSERT INTO plugins (
    id,
    tenant_id,
    manifest,
    plugin_type,
    name,
    version,
    description,
    author_name,
    author_email,
    author_website,
    category,
    status,
    icon_url,
    homepage_url,
    repository_url,
    license,
    size_bytes,
    signature,
    verified,
    config,
    metadata
) VALUES (
    'b2c3d4e5-f6a7-8901-bcde-f23456789012',
    '00000000-0000-0000-0000-000000000001',  -- FunctionFly team ID
    '{
        "name": "github-integration",
        "version": "1.0.0",
        "entry_point": "index.js",
        "runtime": "nodejs",
        "permissions": [
            {"type": "webhook", "action": "send", "resource": "github.com"},
            {"type": "api", "action": "read", "resource": "repositories"},
            {"type": "api", "action": "write", "resource": "issues"},
            {"type": "api", "action": "write", "resource": "pull_requests"},
            {"type": "api", "action": "read", "resource": "workflows"},
            {"type": "api", "action": "trigger", "resource": "workflows"}
        ],
        "hooks": {
            "pre_deploy": "hooks/pre-deploy.js",
            "post_deploy": "hooks/post-deploy.js",
            "on_error": "hooks/on-error.js"
        }
    }'::jsonb,
    'infrastructure',
    'GitHub Integration',
    '1.0.0',
    'Connect your FunctionFly workflows to GitHub repositories. Sync commits, create issues, manage pull requests, and trigger deployments from GitHub Actions webhooks.',
    'FunctionFly Team',
    'plugins@functionfly.com',
    'https://functionfly.com/plugins/github',
    'integration',
    'disabled',  -- disabled until user configures it
    'https://cdn.functionfly.com/plugins/github/icon.svg',
    'https://functionfly.com/plugins/github',
    'https://github.com/functionfly/plugins',
    'MIT',
    128000,  -- size in bytes
    'sig_plugin_github_1.0.0',
    true,  -- verified
    '{}',
    '{"installed_by": "seed", "install_source": "marketplace"}'
) ON CONFLICT (tenant_id, name) DO NOTHING;

-- Create plugin sandbox configuration
INSERT INTO plugin_sandboxes (
    id,
    plugin_id,
    tier,
    cpu_limit,
    memory_limit_mb,
    timeout_seconds,
    network_isolated,
    filesystem_scope,
    max_instances,
    env_vars,
    allowed_domains,
    blocked_domains,
    rate_limit_rpm
) VALUES (
    'c3d4e5f6-a7b8-9012-cdef-a34567890123',
    'b2c3d4e5-f6a7-8901-bcde-f23456789012',
    'worker',  -- sandbox tier
    10,        -- cpu_limit percent
    256,       -- memory_limit_mb
    300,       -- timeout_seconds (5 min)
    false,     -- network_isolated (needs network for API calls)
    'read-only:github-api',  -- filesystem_scope
    3,         -- max_instances
    '{"GITHUB_API_URL": "https://api.github.com", "GITHUB_WEBHOOK_URL": "https://functionfly.com/webhooks/github"}',
    ARRAY['github.com', 'api.github.com'],
    ARRAY['evil.com'],
    60  -- rate_limit_rpm
) ON CONFLICT (plugin_id) DO NOTHING;

-- Create plugin permissions
INSERT INTO plugin_permissions (
    id,
    plugin_id,
    permission_type,
    permission_action,
    resource,
    granted,
    granted_at,
    granted_by,
    expires_at
) VALUES
    ('d4e5f6a7-b8c9-0123-defa-b45678901234', 'b2c3d4e5-f6a7-8901-bcde-f23456789012', 'webhooks', 'send', 'github.com', true, NOW(), NULL, NULL),
    ('e5f6a7b8-c9d0-1234-efab-c56789012345', 'b2c3d4e5-f6a7-8901-bcde-f23456789012', 'api_keys', 'read', 'repositories', true, NOW(), NULL, NULL),
    ('f6a7b8c9-d0e1-2345-fabc-d67890123456', 'b2c3d4e5-f6a7-8901-bcde-f23456789012', 'api_keys', 'write', 'issues', true, NOW(), NULL, NULL),
    ('a7b8c9d0-e1f2-3456-abcd-e78901234567', 'b2c3d4e5-f6a7-8901-bcde-f23456789012', 'api_keys', 'write', 'pull_requests', true, NOW(), NULL, NULL),
    ('b8c9d0e1-f2a3-4567-bcde-f89012345678', 'b2c3d4e5-f6a7-8901-bcde-f23456789012', 'api_keys', 'read', 'workflows', true, NOW(), NULL, NULL),
    ('c9d0e1f2-a3b4-5678-cdef-a90123456789', 'b2c3d4e5-f6a7-8901-bcde-f23456789012', 'api_keys', 'on_write', 'workflows', true, NOW(), NULL, NULL)
ON CONFLICT DO NOTHING;

-- Update marketplace extension with plugin_id reference
UPDATE marketplace_extensions
SET plugin_id = 'b2c3d4e5-f6a7-8901-bcde-f23456789012'
WHERE id = 'a1b2c3d4-e5f6-7890-abcd-ef1234567890';

-- Add first plugin version
INSERT INTO plugin_versions (
    id,
    plugin_id,
    version,
    changelog,
    manifest,
    size_bytes,
    signature,
    release_at
) VALUES (
    '01234567-89ab-cdef-0123-456789abcdef',
    'b2c3d4e5-f6a7-8901-bcde-f23456789012',
    '1.0.0',
    'Initial release with GitHub integration features',
    '{
        "name": "github-integration",
        "version": "1.0.0",
        "entry_point": "index.js",
        "runtime": "nodejs"
    }'::jsonb,
    128000,
    'sig_version_1.0.0',
    NOW()
) ON CONFLICT DO NOTHING;