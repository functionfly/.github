---
title: CLI Reference
description: Complete fly CLI command reference
---

# CLI Reference

Complete reference for the `fly` CLI.

## Commands

### Authentication

```bash
# Login with OAuth
fly login

# Login with dev credentials
fly login --dev

# Logout
fly logout

# Show current user
fly whoami
```

### Project Management

```bash
# Initialize new function
fly init my-function

# List functions
fly list

# View function details
fly info my-function
```

### Development

```bash
# Run locally with hot reload
fly dev

# Run with specific port
fly dev --port 3000

# Test local function
fly test local
```

### Deployment

```bash
# Publish to registry
fly publish

# Build and publish
fly publish --build

# Deploy with options
fly deploy --env production --version 1.0.0
```

### Execution

```bash
# Execute function
fly invoke my-function --data '{"name": "World"}'

# Execute with headers
fly invoke my-function -H "Authorization: Bearer ..."

# Stream execution logs
fly logs my-function --follow
```

### Updates

```bash
# Bump patch version
fly update patch

# Bump minor version
fly update minor

# Bump major version
fly update major
```

### Environment Variables

```bash
# List env vars
fly env list

# Set env var
fly env set API_KEY=xxx

# Get env var
fly env get API_KEY

# Remove env var
fly env unset API_KEY
```

### Secrets

```bash
# List secrets
fly secrets list

# Set secret
fly secrets set DB_PASSWORD=xxx

# Remove secret
fly secrets unset DB_PASSWORD
```

### Monitoring

```bash
# View stats
fly stats my-function

# View health
fly health my-function

# View execution history
fly executions my-function
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
fly login
fly init my-api
fly dev
fly publish
fly logs my-api --follow

# Production deployment
fly deploy --env production --version 2.0.0
fly stats my-api
