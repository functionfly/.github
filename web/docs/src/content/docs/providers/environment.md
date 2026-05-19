---
title: Environment Variables Reference
description: Common environment variables for FunctionFly functions across all providers.
---

This page documents provider-agnostic environment variables available to your functions at runtime.

## Runtime Context Variables

These variables are available to all functions regardless of provider:

| Variable | Description |
|----------|-------------|
| `FUNCTION_NAME` | Name of the executing function |
| `FUNCTION_VERSION` | Version of the executing function |
| `FUNCTION_REGION` | Edge region where the function is executing |
| `FUNCTION_RUNTIME` | Runtime environment (python, nodejs, go, etc.) |
| `FUNCTION_MEMORY_LIMIT_MB` | Memory limit for the function (MB) |
| `FUNCTION_TIMEOUT_MS` | Timeout for the function (milliseconds) |

## Request Context

| Variable | Description |
|----------|-------------|
| `REQUEST_ID` | Unique identifier for the current request |
| `INVOCATION_ID` | Unique ID for this specific invocation |
| `TRACE_ID` | Distributed tracing identifier |

## Platform Information

| Variable | Description |
|----------|-------------|
| `PLATFORM_VERSION` | FunctionFly platform version |
| `PLATFORM_API_URL` | URL for platform API calls |
| `SECRETS_VAULT_URL` | URL for secrets vault service |
| `STATEFABRIC_URL` | URL for StateFabric service |

## Provider Detection

To determine which provider your function is running on:

```javascript
const provider = process.env.FFLY_PROVIDER || 'functionfly-edge';
```

## Common Provider Variables

| Provider | Variable | Description |
|----------|----------|-------------|
| All | `FFLY_PROVIDER` | Provider identifier |
| All | `FFLY_DEPLOYMENT_ID` | Unique deployment ID |
| All | `FFLY_FUNCTION_ID` | Function identifier |
| All | `FFLY_SECRET_KEY` | Provider-specific secret key |

## Secrets Vault

Access secrets from the vault using standard environment variable expansion:

```bash
# In your function code, reference secrets as:
# ${SECRET_NAME} or $SECRET_NAME

# Common patterns:
DATABASE_URL=${VAULT_DB_URL}
API_KEY=${VAULT_API_KEY}
```

## Next Steps

- [FunctionFly Edge environment](./functionfly-edge/environment)
- [Vercel environment](./vercel/environment)
- [Cloudflare Workers environment](./cloudflare-workers/environment)
- [AWS Lambda environment](./aws-lambda/environment)
- [Fly.io environment](./fly-io/environment)
- [Deno Deploy environment](./deno-deploy/environment)