---
title: Vercel Environment
description: Environment variables specific to Vercel deployment.
---

Deploy FunctionFly functions to Vercel's edge network.

## Provider Variables

| Variable | Description |
|----------|-------------|
| `VERCEL` | Set to `1` when running on Vercel |
| `VERCEL_ENV` | Environment (production, preview, development) |
| `VERCEL_REGION` | Region where the function is running |
| `VERCEL_URL` | URL of the deployment |
| `VERCEL_GIT_COMMIT_SHA` | Git commit SHA |
| `VERCEL_GIT_REPOSITORY` | Repository URL |
| `VERCEL_GIT_PROVIDER` | Git provider (github, gitlab, bitbucket) |

## Framework Detection

| Variable | Description |
|----------|-------------|
| `NEXT_PUBLIC_VERCEL_ENV` | Next.js environment variable |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `VERCEL_MAXIMUM_FUNCTION_TIMEOUT` | `10` | Maximum function timeout (seconds) |
| `VERCEL_FUNCTIONmemoryMB` | `1024` | Memory allocation in MB |
| `VERCEL_FUNCTION_CONCURRENCY` | `1` | Concurrency per instance |

## Region Detection

```javascript
const region = process.env.VERCEL_REGION || 'global';
```

## Serverless Function Configuration

```javascript
// next.config.js
module.exports = {
  functions: {
    'api/my-function.js': {
      memory: 1024,
      maxDuration: 10,
    }
  }
}
```

## Environment-Specific Variables

| Environment | Variable | Description |
|-------------|----------|-------------|
| Production | `VERCEL_ENV=production` | Production deployment |
| Preview | `VERCEL_ENV=preview` | Preview deployment |
| Development | `VERCEL_ENV=development` | Local development |

## Secrets

Access Vercel secrets via the dashboard or CLI:

```bash
vercel env pull .env.local
```

FunctionFly secrets vault integrates with Vercel:

```javascript
const apiKey = process.env.VAULT_API_KEY;
```

## Example Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "vercel",
  "environment": {
    "VERCEL_MAXIMUM_FUNCTION_TIMEOUT": "10"
  }
}
```