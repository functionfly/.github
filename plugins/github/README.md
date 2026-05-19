# GitHub Plugin for FunctionFly Studio

## Overview

A first-party GitHub integration plugin created by the FunctionFly team. This plugin connects FunctionFly workflows to GitHub repositories for issue management, PR tracking, webhook handling, and CI/CD automation.

## Files

```
plugins/github/
├── manifest.json         # Plugin manifest with permissions, hooks, config schema
├── index.js             # Main plugin entry point
├── hooks/
│   ├── pre-deploy.js    # Pre-deployment checks
│   ├── post-deploy.js   # Post-deployment status updates
│   └── on-error.js      # Error handling and issue creation
├── lib/
│   └── github-api.js    # GitHub API client
└── endpoints/
    ├── webhook.js       # Webhook handler
    └── handlers.js      # API endpoint handlers
```

## Database Seed

Apply the seed migration to register the plugin:

```bash
psql -h localhost -p 5432 -U postgres -d functionfly -f migrations/seed_github_plugin.up.sql
```

This creates:
- `marketplace_extensions` entry (featured, verified)
- `plugins` entry (disabled until user configures)
- `plugin_sandboxes` configuration (worker tier)
- `plugin_permissions` for all GitHub API scopes
- `plugin_versions` initial version

## Features

### Permissions
- **Webhook**: Send webhook events to GitHub
- **API**: Read/write access to repositories, issues, pull requests
- **API**: Read/trigger GitHub Actions workflows

### Hooks
- `pre_deploy`: Validates GitHub configuration before deployment
- `post_deploy`: Updates commit status and optionally posts PR comments
- `on_error`: Creates GitHub issues for failed workflows

### Configuration
```json
{
  "github_token": "ghp_xxx",        // Required: GitHub PAT
  "webhook_secret": "xxx",          // Required: Webhook signature secret
  "default_repo": "owner/repo",    // Optional: Default repo for actions
  "auto_sync": true                // Optional: Auto-sync on events
}
```

## Installation

1. Go to Studio → Extensions → Plugin Store
2. Find "GitHub Integration" (featured)
3. Click Install
4. Configure your GitHub Personal Access Token
5. Enable the plugin

## Development

To test locally:
```bash
cd plugins/github
node index.js
```

## Sandbox Configuration

- **Tier**: Worker (isolated worker with controlled resources)
- **Memory**: 256MB
- **CPU**: 10%
- **Timeout**: 300 seconds
- **Network**: Allowed to github.com, api.github.com
- **Rate Limit**: 60 requests/minute