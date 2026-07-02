---
title: Model Catalog
description: Full list of available models with capabilities and pricing tiers
sidebar:
  order: 3
---


FunctionFly curates 100+ models from 12 providers. Use the catalog to
choose the right model for your use case.

## OpenAI (Direct)

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `gpt-5.5` | Frontier | 256K | Complex reasoning, analysis |
| `gpt-5.4` | Frontier | 256K | General purpose |
| `gpt-5-mini` | Fast | 128K | High-volume, low-cost |
| `gpt-5-codex` | Code | 256K | Code generation, refactoring |
| `gpt-4.1` | Frontier | 1M | Long-context tasks |
| `gpt-4o` | Frontier | 128K | Multimodal (text + vision) |
| `gpt-4o-mini` | Fast | 128K | Fast, cheap |
| `o3` | Reasoning | 200K | Complex multi-step reasoning |
| `o1` | Reasoning | 200K | Math, science, coding |
| `text-embedding-3-small` | Embedding | — | Vector search (1536 dims) |
| `text-embedding-3-large` | Embedding | — | High-quality embeddings (3072 dims) |

## Anthropic (Direct)

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `claude-opus-4` | Frontier | 200K | Highest quality, complex tasks |
| `claude-sonnet-4.6` | Frontier | 200K | Balanced quality/speed |
| `claude-3.5-sonnet` | Frontier | 200K | General purpose |
| `claude-3.5-haiku` | Fast | 200K | Fast responses, high volume |
| `claude-3-opus` | Frontier | 200K | Legacy, high quality |

## Groq (Ultra-Low Latency)

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `llama-4-scout-17b` | Fast | 128K | Fast inference |
| `llama-4-maverick-17b` | Fast | 1M | Long-context fast inference |
| `llama-3.3-70b` | Fast | 128K | General purpose |
| `qwen-qwq-32b` | Reasoning | 128K | Math, reasoning |
| `deepseek-r1-distill` | Reasoning | 128K | Reasoning at Groq speeds |
| `mixtral-8x7b` | Fast | 32K | Fast general purpose |

## Fireworks

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `llama-3.1-405b` | Frontier | 128K | Largest open model |
| `llama-3.1-70b` | Fast | 128K | Cost-effective |
| `deepseek-v3` | Frontier | 128K | Code + reasoning |
| `deepseek-r1` | Reasoning | 128K | Chain-of-thought |
| `qwen-2.5-coder-32b` | Code | 128K | Code generation |

## DeepInfra (Cost-Optimized)

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `llama-3.3-70b-turbo` | Fast | 128K | Cost-effective general |
| `deepseek-v3` | Frontier | 128K | Code + reasoning |
| `deepseek-r1` | Reasoning | 128K | Chain-of-thought |
| `qwen-2.5-72b` | Fast | 128K | General purpose |
| `bge-large-en-v1.5` | Embedding | — | Embeddings |

## Together AI

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `llama-3.3-70b` | Fast | 128K | General purpose |
| `llama-3.1-405b` | Frontier | 128K | Highest quality open model |
| `deepseek-v3` | Frontier | 128K | Code + reasoning |
| `deepseek-r1` | Reasoning | 128K | Chain-of-thought |
| `mistral-small-24b` | Fast | 32K | Fast, cost-effective |

## MiMo (Xiaomi)

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `mimo-v2.5-pro` | Frontier | 1M | Long-context reasoning |
| `mimo-v2.5-pro-ultraspeed` | Fast | 1M | Fast long-context |
| `mimo-v2.5` | Fast | 128K | General purpose |
| `mimo-ultra` | Frontier | 256K | Highest quality |

## MiniMax

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `minimax-m2.5` | Fast | 256K | Agentic tasks |
| `minimax-m2.7` | Frontier | 512K | Long-context |
| `minimax-m3` | Frontier | 1M | Ultra-long context |
| `minimax-text-01` | Fast | 256K | General text |

## StepFun

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `step-3.5-flash` | Fast | 128K | Fast reasoning |
| `step-3` | Frontier | 128K | General reasoning |
| `step-2-mini` | Fast | 64K | Cost-effective |
| `step-1o-turbo-vision` | Fast | 128K | Vision tasks |

## Ollama (Local)

| Model | Tier | Context | Best For |
|-------|------|---------|----------|
| `llama-3.3` | Local | 128K | Local dev, no API costs |
| `llama-3.2` | Local | 128K | Lighter local dev |
| `qwen-2.5-coder` | Local | 128K | Local code generation |
| `deepseek-r1` | Local | 128K | Local reasoning |
| `codellama` | Local | 16K | Local code |
| `nomic-embed-text` | Embedding | — | Local embeddings |

## OpenRouter (Meta-Gateway)

OpenRouter provides access to 100+ models through a single API key,
including all providers above plus additional models. Use OpenRouter as a
single integration point when you want access to everything.

## Querying the Catalog

Get the full catalog programmatically:

```bash
curl https://api.functionfly.com/v1/ai/models/catalog \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

Filter by provider, tier, or capability:

```bash
curl "https://api.functionfly.com/v1/ai/models/catalog?provider=openai&tier=fast" \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

## Next Steps

- [Bring Your Own Key](/ai-models/byok/) — Connect provider keys
- [Configuration](/ai-models/configuration/) — Profiles and preferences
- [API Reference](/ai-models/api/) — Full endpoint docs
