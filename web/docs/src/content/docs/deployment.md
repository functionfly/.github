---
title: Deployment
description: Deploy your functions to production
---

# Deployment Guide

Deploy your functions to the FunctionFly global registry with automatic scaling and edge distribution.

## Deployment Methods

### 1. CLI Deployment

```bash
# Publish to registry
ffly publish

# Deploy with specific version
ffly deploy --version 1.0.0
```

### 2. CI/CD Integration

```yaml
# .github/workflows/deploy.yml
name: Deploy Function

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: functionfly/action-deploy@v1
        with:
          token: ${{ secrets.FFLY_TOKEN }}
```

## Configuration

### functionfly.jsonc

```jsonc
{
  "name": "my-function",
  "runtime": "python",
  "description": "My awesome function",
  "version": "1.0.0",
  "environment": {
    "NODE_ENV": "production"
  },
  "limits": {
    "timeout": 30,
    "memory": 256
  }
}
```

## Environments

| Environment | URL | Use Case |
|-------------|-----|----------|
| Development | http://localhost:8080 | Local testing |
| Staging | api.staging.functionfly.com | Pre-production |
| Production | api.functionfly.com | Live traffic |

## Deployment Strategies

### Blue-Green Deployment

```bash
# Deploy to staging first
ffly deploy --env staging

# Test and verify
curl https://api.staging.functionfly.com/your-function

# Promote to production
ffly deploy --env production
```

### Canary Releases

```bash
# Deploy canary (10% traffic)
ffly deploy --canary=10

# Increase traffic
ffly deploy --canary=50

# Full rollout
ffly deploy --canary=100
```

## Monitoring

After deployment, monitor your function:

```bash
# View logs
ffly logs my-function

# Check metrics
ffly stats my-function

# Check health
ffly health my-function
```

## Rollback

```bash
# Rollback to previous version
ffly rollback my-function

# Rollback to specific version
ffly rollback my-function --version 1.0.0
```

## Best Practices

1. **Test locally first** - Use `ffly dev` before publishing
2. **Use versions** - Tag releases for easy rollback
3. **Monitor metrics** - Track execution times and errors
4. **Set appropriate limits** - Configure timeout and memory
5. **Use environments** - Test in staging before production
