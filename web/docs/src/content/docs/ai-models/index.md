---
title: AI Models
description: Bring your own AI keys (BYOK) — 100+ models across 12 providers
---


FunctionFly gives you access to **100+ AI models** across **12 providers**.
The recommended approach is to **bring your own AI keys (BYOK)** — connect your
provider keys once and pay the provider directly with no platform markup.

Free OpenRouter models are available as a fallback when no BYOK key is configured.

## Quick Start: Bring Your Own Key (BYOK)

### Connect Your Key

```bash
curl -X POST https://api.functionfly.com/v1/ai-keys/connect \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "api_key": "sk-..."
  }'
```

Your key is validated, encrypted with AES-256-GCM, and stored securely. All AI
calls then use your key at **$0 platform cost**.

### Start Using AI

```bash
curl -X POST https://api.functionfly.com/v1/ai/composer/generate \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a function that summarizes PDFs",
    "model": "gpt-4o"
  }'
```

## Providers

| Provider | Type | Key Models | BYOK |
|----------|------|------------|------|
| **OpenAI** | Frontier + fast | GPT-4o, o1, GPT-5 | ✅ |
| **Anthropic** | Frontier | Claude 3.5 Sonnet, Opus 4 | ✅ |
| **OpenRouter** | Meta-gateway | 100+ models | ✅ |
| **Groq** | Ultra-low latency | Llama 4, QwQ 32B | ✅ |
| **Fireworks** | Structured output | Llama 3.1 405B | ✅ |
| **DeepInfra** | Cost-optimized | Llama 3.3 70B | ✅ |
| **Together AI** | Wide catalog | Llama 3.3 70B | ✅ |
| **MiMo (Xiaomi)** | Long-context | MiMo V2.5 Pro (1M ctx) | ✅ |
| **MiniMax** | Agentic | M2.7 (512K ctx) | ✅ |
| **StepFun** | Reasoning + vision | Step 3.5 Flash | ✅ |
| **Ollama** | Local dev | Llama 3.3, Qwen | ✅ |

## Model Tiers

| Tier | Use Case | Examples |
|------|----------|----------|
| **Frontier** | Highest quality | Claude Opus 4, GPT-4o, MiMo Ultra |
| **Fast** | Low latency, lower cost | GPT-4o Mini, Claude Haiku, Llama 4 Scout |
| **Reasoning** | Chain-of-thought | o3, DeepSeek R1, QwQ 32B |
| **Code** | Code generation | GPT-5 Codex, Qwen3 Coder |
| **Embedding** | Vector embeddings | text-embedding-3-small/large, BGE |
| **Free** | No cost | OpenRouter free models (poolside/laguna-xs.2:free) |

## Model Profiles

FunctionFly ships three pre-configured profiles that select models
automatically based on your use case:

| Profile | Strategy | Example Models |
|---------|----------|----------------|
| **Fast** | Low latency, low cost | Groq Llama 4 Scout, Gemini 2.5 Flash |
| **Balanced** | Default | Claude Sonnet 4.6 via OpenRouter |
| **Premium** | Highest quality | Claude Sonnet 4.6, GPT-4o |

Profiles can be set per-tenant or per-feature (composer, agent chat, FRG, etc.).

## Where Models Are Used

| Feature | Description |
|---------|-------------|
| **AI Composer** | Generate functions from natural language prompts |
| **Agent Chat** | Conversational AI for agents |
| **FRG Assistant** | Visual workflow builder AI |
| **Function DNA** | Evolution analysis |
| **Embeddings** | Vector search and RAG |
| **Support** | AI-powered customer support |

## Plan Limits

| Plan | AI Calls/month | BYOK |
|------|---------------|------|
| Free | 10,000 (rate limits apply) | ✅ Bring your own key |
| Starter | 100,000 | ✅ |
| Professional | 1,000,000 | ✅ |
| Enterprise | 5,000,000 | ✅ |
| Agent Enterprise | Unlimited | ✅ |

BYOK calls are **$0 platform cost** — you pay the provider directly.
Rate limits apply based on your plan tier.

## Platform Fallback (Free Models)

When no BYOK key is configured, FunctionFly uses OpenRouter free models as a
fallback. These are **completely free** but have lower rate limits.

To ensure uninterrupted access, **we recommend connecting your own AI key**.

## Next Steps

- [Bring Your Own Key (BYOK)](/ai-models/byok/) — Connect your provider keys
- [Model Catalog](/ai-models/catalog/) — Full model list with capabilities
- [Configuration](/ai-models/configuration/) — Profiles, preferences, routing
- [API Reference](/ai-models/api/) — Full endpoint documentation
