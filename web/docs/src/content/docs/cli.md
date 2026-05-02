---
title: CLI Reference
description: Complete ffly CLI command reference
---

# CLI Reference

Complete reference for the `fly` CLI.

## Commands

### Authentication

```bash
# Login with OAuth
ffly login

# Login with dev credentials
ffly login --dev

# Logout
ffly logout

# Show current user
ffly whoami
```

### Project Management

```bash
# Initialize new function
ffly init my-function

# List functions
ffly list

# View function details
ffly info my-function
```

### Development

```bash
# Run locally with hot reload
ffly dev

# Run with specific port
ffly dev --port 3000

# Test local function
ffly test local
```

### Deployment

```bash
# Publish to registry
ffly publish

# Build and publish
ffly publish --build

# Deploy with options
ffly deploy --env production --version 1.0.0
```

### Execution

```bash
# Execute function
ffly invoke my-function --data '{"name": "World"}'

# Execute with headers
ffly invoke my-function -H "Authorization: Bearer ..."

# Stream execution logs
ffly logs my-function --follow
```

### Updates

```bash
# Bump patch version
ffly update patch

# Bump minor version
ffly update minor

# Bump major version
ffly update major
```

### Environment Variables

```bash
# List env vars
ffly env list

# Set env var
ffly env set API_KEY=xxx

# Get env var
ffly env get API_KEY

# Remove env var
ffly env unset API_KEY
```

### Secrets

```bash
# List secrets
ffly secrets list

# Set secret
ffly secrets set DB_PASSWORD=xxx

# Remove secret
ffly secrets unset DB_PASSWORD
```

### Monitoring

```bash
# View stats
ffly stats my-function

# View health
ffly health my-function

# View execution history
ffly executions my-function
```

## Configuration

### Config File

The CLI uses `~/.fly/config.yaml`:

```yaml
api_url: https://api.functionfly.com
timeout: 30s
log_level: info
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `FFLY_API_URL` | API base URL |
| `FFLY_TOKEN` | Bearer token |
| `FFLY_CONFIG` | Config file path |

## Options

| Option | Description |
|--------|-------------|
| `--json` | JSON output |
| `--verbose` | Verbose logging |
| `--help` | Show help |

## Examples

```bash
# Full workflow
ffly login
ffly init my-api
ffly dev
ffly publish
ffly logs my-api --follow

# Production deployment
ffly deploy --env production --version 2.0.0
ffly stats my-api
