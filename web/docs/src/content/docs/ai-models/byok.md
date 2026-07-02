---
title: Bring Your Own Key
description: Connect your own provider keys to avoid platform markup
sidebar:
  order: 2
---


BYOK lets you connect your own API keys from supported providers. Calls
made through BYOK keys cost **$0 on FunctionFly** — you pay the provider
directly at their published rates.

## Supported Providers

| Provider | Key Format | Notes |
|----------|-----------|-------|
| `openai` | `sk-...` | All OpenAI models |
| `anthropic` | `sk-ant-...` | All Claude models |
| `openrouter` | `sk-or-...` | 100+ models via OpenRouter |
| `groq` | `gsk_...` | Ultra-low latency inference |
| `fireworks` | `fw_...` | Structured output specialist |
| `deepinfra` | `di_...` | Cost-optimized inference |
| `together` | `tok_...` | Wide model catalog |
| `mimo` | `mimo-...` | Xiaomi MiMo models |
| `mimo-token-plan` | `tp-...` | MiMo prepaid credits |
| `stepfun` | `stf-...` | StepFun reasoning/vision |
| `minimax` | `sk-...` | MiniMax long-context models |
| `minimax-token-plan` | `sk-cp-...` | MiniMax prepaid credits |

## How It Works

```
Your App → FunctionFly API → Decrypt BYOK Key → Provider API → Response
                                   │
                                   ▼
                          AES-256-GCM encrypted
                          in ai_provider_keys table
```

1. **Connect** — You submit your provider API key
2. **Validate** — FunctionFly tests the key against the provider
3. **Encrypt** — Key is encrypted with AES-256-GCM (tenant-specific derivation)
4. **Store** — Encrypted key stored in `ai_provider_keys`
5. **Use** — On each AI call, the key is decrypted and injected transparently

Your plaintext key is **never stored** and **never logged**.

## Connecting a Key

### Dashboard

1. Go to **Settings → AI Keys**
2. Select a provider
3. Paste your API key
4. Click **Connect** — the key is validated and stored

### API

```bash
curl -X POST https://api.functionfly.com/v1/ai-keys/connect \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "api_key": "sk-proj-..."
  }'
```

The key is validated with a lightweight API call before storage.

## Testing a Key

Verify that a connected key still works:

```bash
curl -X POST https://api.functionfly.com/v1/ai-keys/openai/test \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

```json
{
  "valid": true,
  "provider": "openai",
  "tested_at": "2026-06-30T12:00:00Z"
}
```

## Listing Connected Keys

```bash
curl https://api.functionfly.com/v1/ai-keys \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

```json
{
  "keys": [
    {
      "provider": "openai",
      "key_source": "byok",
      "connected_at": "2026-06-01T00:00:00Z",
      "last_used_at": "2026-06-30T12:00:00Z",
      "is_valid": true
    }
  ]
}
```

API keys are never returned in responses — only metadata.

## Rotating a Key

Replace an existing BYOK key with a new one:

```bash
curl -X POST https://api.functionfly.com/v1/ai-keys/openai/rotate \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "sk-proj-new-key..."
  }'
```

## Disconnecting a Key

```bash
curl -X DELETE https://api.functionfly.com/v1/ai-keys/openai \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

After disconnection, calls to that provider fall back to FunctionFly's
managed keys (metered billing).

## Security

- **AES-256-GCM encryption** — Tenant-specific key derivation from `SERVER_MASTER_KEY`
- **No plaintext storage** — Keys are encrypted at rest, decrypted only in memory for the duration of a call
- **No logging** — Plaintext keys are never written to logs
- **Validation on connect** — Keys are tested before storage
- **Rotation support** — Rotate keys without downtime

## Billing

| Key Source | Platform Cost | Provider Cost |
|-----------|--------------|---------------|
| Platform (managed) | 25% markup | Included in FunctionFly bill |
| BYOK | $0 | Paid directly to provider |

BYOK calls still count toward your plan's AI call limit for rate limiting
purposes, but no platform charges are applied.

## Environment Variables (Platform Keys)

If you're running FunctionFly self-hosted, set these environment variables
to provide platform-managed keys:

| Variable | Provider |
|----------|----------|
| `OPENAI_API_KEY` | OpenAI |
| `ANTHROPIC_API_KEY` | Anthropic |
| `OPENROUTER_API_KEY` | OpenRouter |
| `GROQ_API_KEY` | Groq |
| `MIMO_API_KEY` | MiMo |
| `MINIMAX_API_KEY` | MiniMax |
| `STEPFUN_API_KEY` | StepFun |

## Next Steps

- [Model Catalog](/ai-models/catalog/) — Full model list
- [Configuration](/ai-models/configuration/) — Profiles and preferences
- [API Reference](/ai-models/api/) — Full endpoint docs
