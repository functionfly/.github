---
title: Environment Variables
description: Complete reference for all environment variables available in FunctionFly.
sidebar:
  order: 10
---



This page documents all environment variables used by FunctionFly across CLI, API, and runtime environments.

## Core Variables

### API Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `FUNCTIONFLY_API_URL` | Yes | `https://api.functionfly.com` | Base URL for the FunctionFly API |
| `FUNCTIONFLY_API_KEY` | Yes | — | Your API key for authentication |
| `FUNCTIONFLY_API_SECRET` | Yes | — | Shared secret for HMAC request signing |
| `FFLY_API_TIMEOUT` | No | `30000` | API request timeout in milliseconds |

### Authentication

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `FFLY_TOKEN` | No | — | Session token from `ffly login` |
| `FFLY_REFRESH_TOKEN` | No | — | Refresh token for session renewal |
| `FFLY_SESSION_FILE` | No | `~/.ffly/session.json` | Path to session storage file |

### Database Connection

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DB_HOST` | Yes | `localhost` | PostgreSQL host address |
| `DB_PORT` | Yes | `5432` | PostgreSQL port |
| `DB_USER` | Yes | `postgres` | Database username |
| `DB_PASSWORD` | Yes | — | Database password |
| `DB_NAME` | Yes | `functionfly` | Database name |
| `DB_SSLMODE` | No | `require` | PostgreSQL SSL mode (`disable`, `require`, `verify-ca`, `verify-full`) |
| `DB_MAX_CONNECTIONS` | No | `25` | Maximum database connection pool size |
| `DB_IDLE_TIMEOUT_MS` | No | `30000` | Idle connection timeout in milliseconds |

### Redis Cache

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REDIS_ADDR` | Yes | `localhost:6379` | Redis server address |
| `REDIS_PASSWORD` | No | — | Redis password (if authentication enabled) |
| `REDIS_DB` | No | `0` | Redis database number |
| `REDIS_POOL_SIZE` | No | `10` | Maximum Redis connection pool size |
| `REDIS_TIMEOUT_MS` | No | `5000` | Redis operation timeout in milliseconds |

### Application Settings

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | HTTP port for the API server |
| `LOG_LEVEL` | No | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `ENVIRONMENT` | No | `development` | Environment name (`development`, `staging`, `production`) |
| `DEVELOPMENT` | No | `false` | Enable development mode features |
| `BASE_URL` | No | — | Public base URL of your deployment |

---

## Function Runtime Variables

These variables are available to your functions at runtime.

### Request Context

| Variable | Description |
|----------|-------------|
| `REQUEST_ID` | Unique identifier for the current request |
| `FUNCTION_NAME` | Name of the executing function |
| `FUNCTION_VERSION` | Version of the executing function |
| `FUNCTION_MEMORY_LIMIT_MB` | Memory limit for the function (MB) |
| `FUNCTION_TIMEOUT_MS` | Timeout for the function (milliseconds) |
| `FUNCTION_REGION` | Edge region where the function is executing |
| `FUNCTION_RUNTIME` | Runtime environment (python, nodejs, go, etc.) |

### Execution Metadata

| Variable | Description |
|----------|-------------|
| `INVOCATION_ID` | Unique ID for this specific invocation |
| `TRACE_ID` | Distributed tracing identifier |
| `SPAN_ID` | Current span identifier for tracing |
| `START_TIME_MS` | Unix timestamp (milliseconds) when execution started |

### Platform Information

| Variable | Description |
|----------|-------------|
| `PLATFORM_VERSION` | FunctionFly platform version |
| `PLATFORM_API_URL` | URL for platform API calls |
| `SECRETS_VAULT_URL` | URL for secrets vault service |
| `STATEFABRIC_URL` | URL for StateFabric service |

---

## Secrets Vault Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VAULT_ENCRYPTION_KEY` | Yes | — | Master encryption key for secrets vault (256-bit) |
| `VAULT_KEY_ROTATION_DAYS` | No | `90` | Days between automatic key rotation |
| `VAULT_AUDIT_LOG_ENABLED` | No | `true` | Enable audit logging for vault access |

---

## StateFabric Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `STATEFABRIC_ENABLED` | No | `true` | Enable StateFabric state management |
| `STATEFABRIC_CACHE_SIZE_MB` | No | `256` | Hot cache size in megabytes |
| `STATEFABRIC_SNAPSHOT_INTERVAL_SEC` | No | `300` | Snapshot interval in seconds |
| `STATEFABRIC_MAX_OPERATIONS_PER_SEC` | No | `10000` | Rate limit for state operations |
| `STATEFABRIC_REPLICATION_FACTOR` | No | `3` | Number of replicas for durability |

---

## Agent Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AGENTS_ENABLED` | No | `true` | Enable agent functionality |
| `AGENTS_MAX_CONCURRENT` | No | `100` | Maximum concurrent agent executions |
| `AGENTS_MEMORY_LIMIT_MB` | No | `512` | Memory limit per agent |
| `AGENTS_TOOL_CALL_TIMEOUT_MS` | No | `60000` | Timeout for agent tool calls |
| `AGENTS_TRUST_VERIFICATION_ENABLED` | No | `true` | Enable trust verification for agents |
| `AGENTS_LOOKBACK_DAYS` | No | `7` | Days of execution history to retain |

---

## Billing & Wallet Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `WALLET_ENABLED` | No | `true` | Enable wallet and credits system |
| `WALLET_LOW_BALANCE_THRESHOLD_USD` | No | `5.00` | Alert threshold for low balance (USD) |
| `DNA_MUTATION_COST_CREDITS` | No | `50` | Credits charged per DNA mutation acceptance |
| `BILLING_ENABLED` | No | `true` | Enable billing and subscription features |
| `STRIPE_WEBHOOK_SECRET` | No | — | Stripe webhook signing secret |
| `STRIPE_SECRET_KEY` | No | — | Stripe API secret key |

---

## Security Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `HMAC_ENABLED` | No | `true` | Enable HMAC request signing |
| `HMAC_ALGORITHM` | No | `sha256` | HMAC algorithm (`sha256`, `sha512`) |
| `RATE_LIMIT_ENABLED` | No | `true` | Enable rate limiting |
| `RATE_LIMIT_REQUESTS_PER_MINUTE` | No | `1000` | Global requests per minute limit |
| `MALWARE_SCAN_ENABLED` | No | `true` | Enable malware scanning for uploads |
| `TRUST_VERIFICATION_ENABLED` | No | `true` | Enable trust verification (Professional+) |

---

## Data Retention Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATA_RETENTION_ENABLED` | No | `true` | Enable automatic data retention cleanup |
| `DATA_RETENTION_CRON` | No | `0 3 * * *` | Cron schedule for cleanup jobs (default: 3 AM daily) |
| `DATA_RETENTION_DETAILED_DAYS` | No | `90` | Days to keep detailed execution logs |
| `DATA_RETENTION_FINANCIAL_YEARS` | No | `7` | Years to retain financial records (SOX compliance) |
| `DATA_RETENTION_SKIP_IF_LEGAL_HOLD` | No | `true` | Skip cleanup if legal holds are active |

---

## Feature Flags

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `FEATURE_DNA_ENABLED` | No | `true` | Enable Function DNA feature |
| `FEATURE_FRG_ENABLED` | No | `true` | Enable Function Runtime Graph |
| `FEATURE_AGENTS_ENABLED` | No | `true` | Enable AI Agents |
| `FEATURE_BUNDLES_ENABLED` | No | `true` | Enable Backend-in-a-Box bundles |
| `FEATURE_CONSCIOUSNESS_ENABLED` | No | `true` | Enable Function Consciousness (Pro+) |
| `FEATURE_EMBED_ENABLED` | No | `true` | Enable function embedding |

---

## Using Environment Variables in Functions

### Python

```python
import os

api_url = os.environ.get('FUNCTIONFLY_API_URL')
runtime = os.environ.get('FUNCTION_RUNTIME')
region = os.environ.get('FUNCTION_REGION')

def handler(request):
    return {
        "statusCode": 200,
        "body": {"region": region, "runtime": runtime}
    }
```

### JavaScript

```javascript
const apiUrl = process.env.FUNCTIONFLY_API_URL;
const runtime = process.env.FUNCTION_RUNTIME;
const region = process.env.FUNCTION_REGION;

export default async function handler(request) {
    return {
        statusCode: 200,
        body: { region, runtime }
    };
}
```

### Go

```go
package main

import (
    "os"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    apiUrl := os.Getenv("FUNCTIONFLY_API_URL")
    runtime := os.Getenv("FUNCTION_RUNTIME")
    
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"runtime":"` + runtime + `"}`))
}
```

---

## Setting Variables in the CLI

```bash
# Set a variable for all commands
export FUNCTIONFLY_API_KEY=your-api-key

# Or pass inline
FUNCTIONFLY_API_KEY=your-api-key ffly deploy

# Use with .env file
ffly deploy --env-file .env.production
```
