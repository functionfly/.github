# FunctionFly Multi-Provider AI Architecture

This document describes the multi-provider AI inference architecture implemented for FunctionFly's AI service (FlyMind). The architecture routes traffic to optimal providers based on use case, optimizing for latency, cost, and capability.

## Overview

The architecture implements 4 new providers plus the existing OpenRouter for a total of 5 routing options:

| Provider | Best For | Key Features |
|----------|----------|--------------|
| **Fireworks AI** | Structured output, function calling | FireAttention engine, 4x lower latency than vLLM, SOC 2 + HIPAA |
| **Groq** | Real-time low-latency calls | LPU hardware, 0.6-0.9s TTFT consistently, free tier 30 RPM |
| **DeepInfra** | Background/batch tasks | Serverless pricing, up to 90% cost reduction vs provisioned |
| **Together AI** | Alternative/batch processing | 200+ models, batch at 50% less cost |
| **OpenRouter** | Fallback/multi-model routing | 100+ models, single API surface, provider agnostic |

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Traffic-Based Provider Router                    │
├─────────────────────────────────────────────────────────────────────┤
│  Traffic Type          │ Primary Provider    │ Fallback Chain        │
│  ──────────────────────┼─────────────────────┼───────────────────────│
│  Real-time agent calls │ Groq                │ Fireworks → OpenRouter│
│  Structured/JSON/tool  │ Fireworks           │ OpenRouter → Together │
│  Function calling      │ Fireworks           │ OpenRouter → Groq     │
│  Background/batch      │ DeepInfra           │ Together → Fireworks  │
│  Embeddings            │ DeepInfra           │ Fireworks → OpenAI  │
│  General/fallback      │ Fireworks           │ OpenRouter → Together│
└─────────────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Get API Keys

Sign up for the recommended providers:

- **Fireworks AI** (primary): https://fireworks.ai
- **Groq** (real-time): https://groq.com
- **DeepInfra** (batch): https://deepinfra.com
- **Together AI** (alternative): https://together.ai

### 2. Configure Environment

Add to `ai-service/.env`:

```bash
# Primary providers (recommended)
FIREWORKS_API_KEY=fw-xxx
GROQ_API_KEY=gsk-xxx
DEEPINFRA_API_KEY=xxx
TOGETHER_API_KEY=xxx

# Set Fireworks as default for function calling
DEFAULT_PROVIDER=fireworks

# Enable traffic-based routing
ENABLE_TRAFFIC_BASED_ROUTING=true
```

### 3. Run the Service

```bash
cd ai-service
source .env
uv sync
PYTHONPATH=. uv run uvicorn src.main:app --host 127.0.0.1 --port 8081
```

## Provider Details

### Fireworks AI (`fireworks`)

**Best for:** Function calling, structured output, JSON mode

**Models:**
- `accounts/fireworks/models/llama-v3p1-405b-instruct` (default)
- `accounts/fireworks/models/llama-v3p1-70b-instruct`
- `accounts/fireworks/models/qwen2p5-72b-instruct`
- `accounts/fireworks/models/deepseek-v3`

**Features:**
- FireAttention engine (4x faster than vLLM)
- Native function calling support
- SOC 2 + HIPAA compliant
- 15+ global regions

**Pricing:**
- Llama 3.1 405B: $3/1M input, $3/1M output
- Llama 3.1 70B: $0.9/1M input, $0.9/1M output
- Llama 3.1 8B: $0.2/1M input, $0.2/1M output

**Configuration:**
```bash
FIREWORKS_API_KEY=fw-xxx
FIREWORKS_RATE_LIMIT=120
```

### Groq (`groq`)

**Best for:** Real-time agent function calls, user-facing interactions

**Models:**
- `llama-4-scout-17b-16e-instruct` (default)
- `llama-4-maverick-17b-128k-instruct`
- `llama-3.3-70b-versatile`
- `qwen-2.5-32b`
- `deepseek-r1-distill-llama-70b`

**Features:**
- LPU (Language Processing Unit) hardware
- 0.6-0.9s time-to-first-token consistently
- Free tier: 30 RPM
- Paid tier: 3000+ RPM

**Limitations:**
- No embeddings support
- Lower rate limits on free tier

**Configuration:**
```bash
GROQ_API_KEY=gsk-xxx
GROQ_RATE_LIMIT=30
```

### DeepInfra (`deepinfra`)

**Best for:** Background tasks, batch processing, embeddings

**Models:**
- `meta-llama/Llama-3.3-70B-Instruct-Turbo` (default)
- `meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo`
- `deepseek-ai/DeepSeek-V3`
- `Qwen/Qwen2.5-72B-Instruct`

**Embeddings:**
- `BAAI/bge-large-en-v1.5` (default)
- `BAAI/bge-base-en-v1.5`
- `nomic-ai/nomic-embed-text-v1`

**Features:**
- Serverless pricing (up to 90% cheaper)
- Batch processing support
- Embedding models available
- Good throughput for background work

**Pricing:**
- Llama 3.3 70B: $0.30/1M input, $0.50/1M output
- Llama 3.1 8B: $0.06/1M input, $0.10/1M output
- Embeddings: $0.02/1M tokens

**Configuration:**
```bash
DEEPINFRA_API_KEY=xxx
DEEPINFRA_RATE_LIMIT=100
DEEPINFRA_EMBEDDING_MODEL=BAAI/bge-large-en-v1.5
```

### Together AI (`together`)

**Best for:** Alternative inference, batch at lower cost

**Models:**
- `meta-llama/Llama-3.3-70B-Instruct-Turbo` (default)
- `meta-llama/Llama-3.2-90B-Vision-Instruct`
- `deepseek-ai/DeepSeek-V3`
- `microsoft/phi-4`

**Embeddings:**
- `BAAI/bge-large-en-v1.5`
- `togethercomputer/m2-bert-80M-2k-retrieval`

**Features:**
- Up to 2x faster serverless inference
- Batch processing at 50% less cost
- 200+ model catalog

**Configuration:**
```bash
TOGETHER_API_KEY=xxx
TOGETHER_RATE_LIMIT=60
```

## Traffic-Based Routing

The router automatically classifies traffic and routes to the optimal provider:

```python
from src.models.schemas import CompletionRequest, TrafficType
from src.providers.router import get_provider_router

router = get_provider_router()

# Route a request with automatic classification
provider, provider_type, model_override = await router.route_completion(
    request,
    traffic_hint=None  # Auto-classified
)

# Or specify explicit traffic type
provider, provider_type, _ = await router.route_completion(
    request,
    traffic_hint=TrafficType.REALTIME
)
```

### Classification Rules

1. **Function Calling**: Messages contain "function", "tool", "invoke", "parameters"
2. **Structured/JSON**: Stop sequences contain "json", "schema"
3. **Real-time**: Short messages (≤2 messages, <200 chars)
4. **Background**: All embedding requests, bulk operations
5. **General**: Default for everything else

### Customizing Routing

Update routing rules at runtime:

```python
from src.models.schemas import ProviderType, TrafficType

router.update_rule(
    traffic_type=TrafficType.REALTIME,
    primary_provider=ProviderType.FIREWORKS,  # Change primary
    fallback_providers=[ProviderType.GROQ, ProviderType.OPENROUTER],
)
```

## Configuration Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `FIREWORKS_API_KEY` | - | Fireworks AI API key |
| `GROQ_API_KEY` | - | Groq API key |
| `DEEPINFRA_API_KEY` | - | DeepInfra API key |
| `TOGETHER_API_KEY` | - | Together AI API key |
| `DEFAULT_PROVIDER` | `fireworks` | Default LLM provider |
| `DEFAULT_EMBEDDING_PROVIDER` | `deepinfra` | Default for embeddings |
| `ENABLE_TRAFFIC_BASED_ROUTING` | `true` | Enable smart routing |
| `TRAFFIC_REALTIME_PROVIDER` | `groq` | Provider for real-time |
| `TRAFFIC_STRUCTURED_PROVIDER` | `fireworks` | Provider for structured |
| `TRAFFIC_FUNCTION_CALLING_PROVIDER` | `fireworks` | Provider for function calls |
| `TRAFFIC_BACKGROUND_PROVIDER` | `deepinfra` | Provider for background |
| `TRAFFIC_FALLBACK_PROVIDER` | `openrouter` | Fallback provider |

### Provider Manager API

```python
from src.providers.manager import get_provider_manager

manager = get_provider_manager()

# Get provider for traffic type
from src.providers.router import get_provider_router
router = manager.get_provider_router()

# Get recommendations
recommendations = manager.get_functionfly_recommendations()
```

## Cost Comparison

Estimated costs per 1M tokens (input/output):

| Provider | 70B Model | 8B Model | Embeddings |
|----------|-----------|----------|------------|
| Fireworks | $0.90/$0.90 | $0.20/$0.20 | $0.02 |
| Groq | $0.59/$0.79 | $0.05/$0.08 | N/A |
| DeepInfra | $0.30/$0.50 | $0.06/$0.10 | $0.02 |
| Together | $0.88/$0.88 | $0.18/$0.18 | $0.02 |
| OpenAI | $0.60/$0.90 | $0.15/$0.20 | $0.02 |

## Migration Guide

### From Single Provider

**Before:**
```bash
DEFAULT_PROVIDER=openai
OPENAI_API_KEY=sk-xxx
```

**After:**
```bash
# Add new providers
FIREWORKS_API_KEY=fw-xxx
GROQ_API_KEY=gsk-xxx
DEEPINFRA_API_KEY=xxx

# Enable routing
DEFAULT_PROVIDER=fireworks
ENABLE_TRAFFIC_BASED_ROUTING=true
```

### No Code Changes Required

All providers are OpenAI-compatible. Your existing code works without changes:

```python
# This works with any provider
response = await provider.complete(
    messages=[...],
    model="llama-v3p1-405b-instruct",
    temperature=0.7,
)
```

## Troubleshooting

### Provider Not Available

Check logs for initialization errors:
```
WARNING: Failed to initialize Fireworks AI provider: FIREWORKS_API_KEY not set
```

### Fallback Behavior

If a provider fails, the router automatically tries fallbacks. To disable:
```bash
ENABLE_TRAFFIC_BASED_ROUTING=false
```

### Rate Limiting

Adjust rate limits based on your quotas:
```bash
GROQ_RATE_LIMIT=30  # Free tier
GROQ_RATE_LIMIT=3000  # Paid tier
```

## API Endpoints

The provider status can be checked at:

```
GET /v1/providers/status
```

Response includes availability and routing info for all providers.

## License

Same as FunctionFly project.
