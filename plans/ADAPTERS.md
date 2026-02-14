# Provider Adapters (MVP1)

MVP1 supports adapters for:

- Cloudflare Workers
- Vercel
- Fly.io
- Deno Deploy

## Safest default: BYO (no stored provider tokens)

MVP1 default is strict BYO:

- Customer deploys the Edge Target into each provider.
- Customer registers the resulting endpoint URL in FunctionFly.
- FunctionFly performs health checks and routes traffic.

Tradeoffs:

- Pros: minimal security surface, no credential storage.
- Cons: less automated; can add opt-in token storage later.

## Common Edge Target contract

All providers implement the same HTTP contract.

### GET /healthz

- returns 200 when healthy
- includes optional JSON payload describing region/build

### GET /ping

- returns 200 quickly
- used for latency probes

### /* (proxy target)

- handles the application’s real request paths

### Request signing

- orchestrator sets `X-FFLY-Timestamp` and `X-FFLY-Signature`
- signature is `HMAC_SHA256(sharedSecret, timestamp + method + path + bodyHash)`

Edge Target verifies signature.

## Provider notes

### Cloudflare Workers

- Good for: global presence and built-in cache.
- Deployment artifact: Worker script.

### Vercel

- Use Edge Functions when possible for latency; fall back to Serverless if needed.
- Ensure `Cache-Control` semantics are explicit.

### Fly.io

- Customer deploys a tiny service (e.g., Go) that implements Edge Target.
- Regions mapped to Fly regions.

### Deno Deploy

- Deno Deploy project exposes Edge Target endpoints.

