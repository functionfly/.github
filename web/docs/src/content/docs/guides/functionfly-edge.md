---
title: FunctionFly Edge
description: Deploy FunctionFly functions to the FunctionFly global edge network
---

# FunctionFly Edge

Deploy your FunctionFly functions to the FunctionFly global edge network for the fastest possible execution with deep platform integration.

## Features

- **Fastest cold start** - Sub-millisecond execution
- **35+ edge locations** - Global anycast network
- **Zero-configuration** - Automatic scaling and distribution
- **Deep platform integration** - Trust API, Secrets Vault, StateFabric

## Prerequisites

- FunctionFly account (sign up at [functionfly.com](https://functionfly.com))
- FunctionFly CLI installed

## Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "functionfly-edge",
  "provider_config": {
    "url": "https://edge.functionfly.com",
    "wasm_url": "https://wasm.functionfly.com",
    "regions": ["global"]
  }
}
```

## Deployment

```bash
# Deploy to FunctionFly Edge (default)
ffly deploy

# Explicitly deploy to FunctionFly Edge
ffly deploy --provider functionfly-edge

# Deploy to specific region
ffly deploy --provider functionfly-edge --region us-east
```

## Environment Variables

```bash
# FunctionFly uses your account credentials
ffly login
ffly whoami
```

## Available Regions

| Region | Location |
|--------|----------|
| global | Anycast (automatic) |
| us-east | US East (Virginia) |
| us-west | US West (California) |
| eu-west | EU West (Ireland) |
| eu-central | EU Central (Frankfurt) |
| ap-east | Asia Pacific (Hong Kong) |
| ap-south | Asia Pacific (Singapore) |

## WASM Runtime

FunctionFly Edge supports WASM for additional isolation and portability:

```jsonc
{
  "provider": "functionfly-wasm",
  "runtime": "wasm",
  "provider_config": {
    "wasm_url": "https://wasm.functionfly.com"
  }
}
```

## Features Only on FunctionFly Edge

- **Trust API** - Built-in verification and attestation
- **Secrets Vault** - Zero-knowledge secret management
- **StateFabric** - Durable state at the edge
- **Native WASM** - First-class WASM support
- **Function Registry** - Built-in function marketplace

## Limitations

- Maximum 512MB memory
- Maximum 30 second timeout
- Requires FunctionFly account