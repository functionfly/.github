---
title: Fly.io Environment
description: Environment variables specific to Fly.io deployment.
---

Deploy FunctionFly functions to Fly.io's distributed platform.

## Provider Variables

| Variable | Description |
|----------|-------------|
| `FLY_APP_NAME` | Application name |
| `FLY_APP_ID` | Application ID |
| `FLY_REGION` | Primary region |
| `FLY_MACHINE_ID` | Machine identifier |
| `FLY_MACHINE_LAUNCH_ID` | Machine launch ID |
| `FLY_IMAGE_REF` | Docker image reference |
| `FLY_PRIMARY_REGION` | Primary region |
| `FLY_VOLUME_MOUNT_POINT` | Volume mount path |
| `FLY_ALLOC_ID` | Allocation ID |

## Regions

Available regions include:

| Region | Code |
|--------|------|
| Amsterdam | `ams` |
| Dallas | `dfw` |
| Frankfurt | `fra` |
| Hong Kong | `hkg` |
| London | `lhr` |
| New York | `ewr` |
| San Jose | `sjc` |
| Seattle | `sea` |
| Singapore | `sin` |
| Tokyo | `nrt` |

## Persistent Storage

| Variable | Description |
|----------|-------------|
| `FLY_VOLUME_PATH` | Path to persistent volume |
| `FLY_VOLUME_SIZE_GB` | Volume size in GB |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `FLY_MEMORY_MB` | `256` | Memory allocation |
| `FLY_CPU_COUNT` | `1` | Number of CPUs |
| `FLY_GLOBAL_SECONDS` | `600` | Seconds per Global second |

## Health Checks

| Variable | Default | Description |
|----------|---------|-------------|
| `FLY_HEALTHCHECK_INTERVAL_SEC` | `30` | Health check interval |
| `FLY_HEALTHCHECK_TIMEOUT_SEC` | `5` | Health check timeout |
| `FLY_HEALTHCHECK_PATH` | `/health` | Health check path |

## Secrets

Store secrets in Fly.io secrets:

```bash
fly secrets set API_KEY=your-secret
fly secrets set DATABASE_URL=postgres://...
```

## Example Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "fly",
  "environment": {
    "FLY_MEMORY_MB": "512",
    "FLY_CPU_COUNT": "1"
  },
  "mount": {
    "/data": {
      "size_gb": 10
    }
  }
}
```

## Scale Configuration

```bash
# Scale to 3 instances
fly scale count 3 --region ams

# Scale with specific resources
fly scale vm shared-cpu-1x --memory 512
```