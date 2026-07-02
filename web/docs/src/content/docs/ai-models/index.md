---
title: AI Models
description: 100+ models across 12 providers, with bring-your-own-key support
---


FunctionFly gives you access to **100+ AI models** across **12 providers**
through a single platform. Use FunctionFly's managed keys (metered billing)
or connect your own provider keys (BYOK) to pay the provider directly.

## Providers

| Provider | Type | Key Models |
|----------|------|------------|
| **OpenRouter** | Meta-gateway | Claude Opus 4.7, Sonnet 4.6, GPT-5.5, Gemini 3.1 Pro, DeepSeek V4, Grok 4, Llama 3.3, o3 |
| **OpenAI** | Frontier + fast | GPT-5.5, GPT-5 Mini, GPT-5 Codex, GPT-4.1, GPT-4o, o3, o1 |
| **Anthropic** | Frontier | Claude Opus 4, Sonnet 4.6, 3.5 Sonnet, 3.5 Haiku |
| **Groq** | Ultra-low latency | Llama 4 Scout/Maverick, QwQ 32B, DeepSeek R1 Distill, Mixtral 8x7B |
| **Fireworks** | Structured output | Llama 3.1 405B/70B, DeepSeek V3/R1, Qwen 2.5 Coder |
| **DeepInfra** | Cost-optimized | Llama 3.3 70B, DeepSeek V3/R1, Qwen 2.5 72B, BGE embeddings |
| **Together AI** | Wide catalog | Llama 3.3 70B, Llama 3.1 405B, DeepSeek V3/R1, Mistral Small |
| **MiMo (Xiaomi)** | Long-context reasoning | MiMo V2.5 Pro (1M context), UltraSpeed, MiMo Ultra |
| **MiniMax** | Agentic/long context | M2.5 (256K), M2.7 (512K), M3 (1M context) |
| **StepFun** | Reasoning + vision | Step 3.5 Flash, Step 3, Step 1o Turbo Vision |
| **Ollama** | Local dev | Llama 3.3, Qwen 2.5 Coder, DeepSeek R1, Code Llama |

## Model Tiers

| Tier | Use Case | Examples |
|------|----------|----------|
| **Frontier** | Highest quality | Claude Opus 4, GPT-5.5, MiMo Ultra |
| **Fast** | Low latency, lower cost | GPT-5 Mini, Claude Haiku, Llama 4 Scout |
| **Reasoning** | Chain-of-thought | o3, DeepSeek R1, MiMo V2.5 Pro |
| **Code** | Code generation | GPT-5 Codex, Qwen3 Coder |
| **Embedding** | Vector embeddings | text-embedding-3-small/large, BGE |
| **Local** | Free, self-hosted | Ollama models |

## Model Profiles

FunctionFly ships three pre-configured profiles that select models
automatically based on your use case:

| Profile | Strategy | Example Models |
|---------|----------|---------------|
| **Fast** | Low latency, low cost | Groq Llama 4 Scout, Gemini 2.5 Flash |
| **Balanced** | Default | Claude Sonnet 4.6 via OpenRouter |
| **Premium** | Highest quality | Claude Sonnet 4.6, Gemini 3.1 Pro |

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

| Plan | AI Calls/month | Price |
|------|---------------|-------|
| Free | 10,000 | $0 |
| Starter | 100,000 | $24/mo |
| Professional | 1,000,000 | $79/mo |
| Enterprise | 5,000,000 (included) | $299/mo |
| Agent Enterprise | Unlimited | $499/mo |

Platform-managed calls carry a **25% markup** on provider costs. BYOK calls
are **$0 platform cost** — you pay the provider directly.

## Quick Start

### Use Platform Keys (Default)

No setup needed. Call any model through the platform:

```bash
curl -X POST https://api.functionfly.com/v1/ai/composer/generate \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a function that summarizes PDFs",
    "model": "claude-sonnet-4-6"
  }'
```

### Connect Your Own Key (BYOK)

```bash
curl -X POST https://api.functionfly.com/v1/ai-keys/connect \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "api_key": "sk-..."
  }'
```

Your key is validated, encrypted with AES-256-GCM, and stored. All subsequent
calls to OpenAI models use your key at **$0 platform cost**.

## Next Steps

- [Bring Your Own Key](/ai-models/byok/) — Connect your provider keys
- [Model Catalog](/ai-models/catalog/) — Full model list with capabilities
- [Configuration](/ai-models/configuration/) — Profiles, preferences, routing
- [API Reference](/ai-models/api/) — Full endpoint documentation
