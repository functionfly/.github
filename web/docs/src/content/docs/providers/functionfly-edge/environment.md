---
title: FunctionFly Edge Environment
description: Environment variables specific to FunctionFly Edge deployment.
---

FunctionFly Edge is the built-in provider with zero configuration required.

## Provider Variables

| Variable | Description |
|----------|-------------|
| `FFLY_EDGE_NODE_ID` | Unique identifier for the edge node |
| `FFLY_EDGE_REGION` | Physical region where function is running |
| `FFLY_EDGE_ZONE` | Availability zone within the region |
| `FFLY_EDGE_DC` | Data center identifier |

## Regions

| Variable | Description |
|----------|-------------|
| `FFLY_EDGE_POP_US_EAST` | US East point of presence |
| `FFLY_EDGE_POP_US_WEST` | US West point of presence |
| `FFLY_EDGE_POP_EU_WEST` | EU West point of presence |
| `FFLY_EDGE_POP_EU_CENTRAL` | EU Central point of presence |
| `FFLY_EDGE_POP_AP_SOUTH` | Asia Pacific South point of presence |
| `FFLY_EDGE_POP_AP_NORTHEAST` | Asia Pacific Northeast point of presence |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `FFLY_EDGE_COLD_START_MS` | `0` | Expected cold start time (ms) |
| `FFLY_EDGE_MAX_MEMORY_MB` | `512` | Maximum memory allocation |
| `FFLY_EDGE_REQUEST_TIMEOUT_MS` | `30000` | Request timeout in milliseconds |

## Stateless Mode

Enable stateless execution for pure function workloads:

```bash
FFLY_EDGE_STATELESS=true
```

## Trust API Integration

When Trust API is enabled:

| Variable | Description |
|----------|-------------|
| `FFLY_TRUST_SCORE` | Trust score for the current invocation |
| `FFLY_TRUST_LEVEL` | Trust level (none, basic, verified, trusted) |
| `FFLY_TRUST_TIMESTAMP` | Timestamp of trust verification |

## StateFabric Integration

StateFabric is enabled by default on FunctionFly Edge:

| Variable | Default | Description |
|----------|---------|-------------|
| `STATEFABRIC_ENABLED` | `true` | Enable StateFabric state management |
| `STATEFABRIC_CACHE_SIZE_MB` | `256` | Hot cache size in megabytes |
| `STATEFABRIC_REPLICATION_FACTOR` | `3` | Number of replicas |

## Secrets Vault

Secrets vault is automatically available:

```javascript
const dbPassword = process.env.VAULT_DB_PASSWORD;
```

## Example Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "functionfly-edge",
  "environment": {
    "FFLY_EDGE_COLD_START_MS": "0",
    "STATEFABRIC_CACHE_SIZE_MB": "512"
  }
}
```