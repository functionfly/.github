---
title: Providers
description: Deploy FunctionFly functions to your preferred cloud provider
---

FunctionFly supports deploying your functions to multiple cloud providers. Each provider offers unique features, regional coverage, and pricing models.

## Supported Providers

| Provider | ID | Description | Environment |
|----------|-----|-------------|-------------|
| [FunctionFly Edge](./functionfly-edge/environment) | `functionfly-edge` | Built-in global edge network with Trust API, Secrets Vault, StateFabric | [Docs](./functionfly-edge/environment) |
| [Vercel](./vercel/environment) | `vercel` | Edge functions with global CDN, Next.js integration | [Docs](./vercel/environment) |
| [Cloudflare Workers](./cloudflare-workers/environment) | `workers` | Ultra-low latency edge compute with Workers KV | [Docs](./cloudflare-workers/environment) |
| [Fly.io](./fly-io/environment) | `fly` | Distributed apps with persistent volumes | [Docs](./fly-io/environment) |
| [AWS Lambda](./aws-lambda/environment) | `aws-lambda` | Enterprise-grade serverless, deep AWS integration | [Docs](./aws-lambda/environment) |
| [Deno Deploy](./deno-deploy/environment) | `deno-deploy` | Native TypeScript runtime with Deno KV | [Docs](./deno-deploy/environment) |

## Quick Comparison

| Best For | Provider |
|----------|----------|
| Fastest cold start | FunctionFly Edge (<1ms) |
| Most memory | AWS Lambda (10GB) |
| Longest timeout | AWS Lambda (900s) |
| Zero config | FunctionFly Edge |
| Next.js projects | Vercel |
| KV storage | Cloudflare Workers, Deno Deploy |
| Persistent storage | Fly.io, AWS Lambda |
| Enterprise scale | AWS Lambda |

## Multi-Provider Deployment

Deploy to multiple providers simultaneously for redundancy or geographic distribution:

```jsonc
// functionfly.jsonc
{
  "providers": ["functionfly-edge", "vercel", "workers"],
  "regions": ["us-east", "eu-west", "ap-south"]
}
```

## Getting Started

1. **FunctionFly Edge**: No setup required - just run `ffly deploy`
2. **Other providers**: Connect your account in the dashboard under Settings → Providers
3. **Deploy**: Use `ffly deploy --provider <provider-id>` or set in config

See individual provider guides for detailed setup instructions.