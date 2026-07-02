---
title: Configuration
description: Model profiles, per-tenant preferences, and routing strategies
sidebar:
  order: 4
---


FunctionFly provides multiple layers of model configuration — from global
profiles to per-feature model overrides.

## Model Profiles

Profiles are pre-configured model selections optimized for different goals:

| Profile | Goal | Example Selection |
|---------|------|-------------------|
| **Fast** | Low latency, low cost | Groq Llama 4 Scout, Gemini 2.5 Flash |
| **Balanced** | Default quality/speed | Claude Sonnet 4.6 via OpenRouter |
| **Premium** | Highest quality | Claude Sonnet 4.6, Gemini 3.1 Pro |

### Setting a Profile

**Dashboard:** Settings → AI Models → Profile

**API:**

```bash
curl -X PUT https://api.functionfly.com/v1/ai/models/preferences \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "profile": "premium"
  }'
```

## Per-Tenant Preferences

Tenant admins can configure AI behavior across the organization:

```bash
curl -X PUT https://api.functionfly.com/v1/ai/models/preferences \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "profile": "balanced",
    "feature_overrides": {
      "composer": "gpt-5-codex",
      "agent": "claude-sonnet-4-6",
      "embeddings": "text-embedding-3-large"
    },
    "allowed_providers": ["openai", "anthropic", "groq"],
    "allowed_models": ["gpt-5.5", "claude-sonnet-4.6", "llama-4-scout"],
    "allow_user_overrides": true
  }'
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `profile` | string | `fast`, `balanced`, or `premium` |
| `feature_overrides` | object | Model per feature (see below) |
| `allowed_providers` | string[] | Restrict to specific providers |
| `allowed_models` | string[] | Restrict to specific models |
| `allow_user_overrides` | bool | Let users choose their own models |

### Feature Keys

| Feature | Description |
|---------|-------------|
| `composer` | AI function generation |
| `agent` | Agent chat/reasoning |
| `frg` | FRG visual workflow assistant |
| `chat` | General AI chat |
| `support` | Customer support AI |
| `dna` | DNA/evolution analysis |
| `embeddings` | Vector embeddings |

## Routing Strategies

When no specific model is set, FunctionFly selects a model using the
configured routing strategy:

| Strategy | Description |
|----------|-------------|
| `quality_first` | Pick the highest-quality available model (default) |
| `balanced` | Balance quality and cost |
| `cost_optimized` | Prefer cheaper models |
| `cost_first` | Always pick the cheapest option |

## Per-Feature Model Selection

Set different models for different features. For example, use a fast model
for embeddings and a premium model for code generation:

```json
{
  "feature_overrides": {
    "composer": "gpt-5-codex",
    "agent": "claude-opus-4",
    "embeddings": "text-embedding-3-small",
    "chat": "gpt-5-mini"
  }
}
```

## User-Level Overrides

If `allow_user_overrides` is enabled in tenant preferences, individual users
can set their own model preferences. User overrides take precedence over
tenant preferences for that user's sessions.

## Testing Model Availability

Check if a model is available on its provider:

```bash
curl -X POST https://api.functionfly.com/v1/ai/models/check \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-6",
    "provider": "anthropic"
  }'
```

```json
{
  "available": true,
  "model": "claude-sonnet-4-6",
  "provider": "anthropic",
  "latency_ms": 245
}
```

## Refreshing the Catalog

Force a refresh of the cached model catalog:

```bash
curl -X POST https://api.functionfly.com/v1/ai/models/catalog/refresh \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

## Production Recommendations

### High-Traffic Applications

- Use **Groq** or **DeepInfra** for cost-optimized inference
- Set `cost_optimized` routing strategy
- Use BYOK to avoid platform markup on high volumes
- Set per-key rate limits to control spend

### Code Generation

- Use `gpt-5-codex` or `qwen-2.5-coder-32b` for the composer
- Set `feature_overrides.composer` to your preferred code model

### Agent Workloads

- Use Claude Sonnet 4.6 or GPT-5.5 for reasoning-heavy agents
- Use Claude Haiku or GPT-5 Mini for high-volume agent chat
- Consider Groq for agents that need ultra-low latency

### Cost Control

- Connect BYOK keys for providers you use heavily
- Set `allowed_models` to restrict to known-cost models
- Use `cost_first` routing for non-critical workloads
- Monitor usage per model in the dashboard

## Next Steps

- [Bring Your Own Key](/ai-models/byok/) — Connect provider keys
- [Model Catalog](/ai-models/catalog/) — Full model list
- [API Reference](/ai-models/api/) — Full endpoint docs
