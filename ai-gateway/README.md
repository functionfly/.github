# AI Gateway

FastAPI service for unified AI inference routing across multiple backends.

## Overview

AI Gateway provides a unified interface for AI inference requests, supporting:

- **RunPod GPU Clusters**: Self-hosted models on GPU instances
- **ONNX Runtime**: Local inference with optimized ONNX models
- **OpenAI-Compatible API**: Integration with OpenAI and OpenRouter

## Architecture

```
┌─────────────────┐      ┌──────────────────────────────────────┐
│  WASM Runtime   │─────▶│           AI Gateway                 │
│  ai_infer call  │      │  ┌─────────────┐  ┌──────────────┐  │
└─────────────────┘      │  │ /infer      │  │ Circuit      │  │
                         │  │ /infer/batch│  │ Breaker      │  │
                         │  └─────────────┘  └──────────────┘  │
                         │         │                  │        │
                         │         ▼                  ▼        │
                         │  ┌─────────────────────────────────┐│
                         │  │      Inference Engine          ││
                         │  │  ┌─────────┐ ┌─────────────┐  ││
                         │  │  │ RunPod  │ │ ONNX Runtime│  ││
                         │  │  │ API     │ │             │  ││
                         │  │  └─────────┘ └─────────────┘  ││
                         │  └─────────────────────────────────┘│
                         └──────────────────────────────────────┘
                                        │
                                        ▼
                         ┌──────────────────────────────────────┐
                         │      RunPod Cluster Manager          │
                         │      (Go service on :8080)           │
                         └──────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
             ┌────────────┐      ┌────────────┐      ┌────────────┐
             │us-east-1   │      │eu-west-1   │      │ap-southeast-│
             │A100 Pool   │      │A100 Pool   │      │1 A100 Pool  │
             └────────────┘      └────────────┘      └────────────┘
```

## Quick Start

### Installation

```bash
# Install with basic dependencies
pip install ai-gateway

# Install with ONNX support
pip install ai-gateway[onnx]

# Install with GPU support
pip install ai-gateway[gpu]

# Install for development
pip install -e ai-gateway[dev]
```

### Configuration

Create a `.env` file:

```env
# Server
HOST=0.0.0.0
PORT=8082

# RunPod Integration
RUNPOD_API_KEY=your_api_key
RUNPOD_CLUSTER_URL=http://localhost:8080

# Model Settings
DEFAULT_MODEL=phi-3-mini
MAX_CONTEXT_LENGTH=4096

# Security
API_KEY_HEADER=X-API-Key
REQUIRED_API_KEY=your_static_api_key

# Performance
MAX_BATCH_SIZE=8
BATCH_TIMEOUT_MS=100
```

### Running

```bash
# Start the server
python -m src.main

# Or with uvicorn directly
uvicorn src.main:app --host 0.0.0.0 --port 8082
```

### Docker

```bash
# Build image
docker build -t ai-gateway:latest .

# Run container
docker run -p 8082:8082 \
  -e RUNPOD_API_KEY=your_key \
  -e RUNPOD_CLUSTER_URL=http://host.docker.internal:8080 \
  ai-gateway:latest
```

## API Endpoints

### Inference

#### POST /v1/infer

Run inference request.

```bash
curl -X POST http://localhost:8082/v1/infer \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_api_key" \
  -d '{
    "model": "onnx://phi-3-mini",
    "input": "SGVsbG8gV29ybGQh",  # base64 encoded
    "parameters": {
      "temperature": 0.7,
      "max_tokens": 100
    }
  }'
```

Response:

```json
{
  "output": "SGVsbG8gZnJvbSBhaS1nYXRld2F5",  # base64 encoded
  "latency_ms": 45.5,
  "cost_usd": 0.002,
  "model": "onnx://phi-3-mini",
  "provider": "onnx",
  "backend": "onnx_runtime",
  "tokens_generated": 50,
  "region": "us-east-1",
  "request_id": "uuid-here",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### POST /v1/infer/batch

Run batch inference (max 8 requests).

```bash
curl -X POST http://localhost:8082/v1/infer/batch \
  -H "Content-Type: application/json" \
  -d '{
    "requests": [
      {"model": "onnx://phi-3-mini", "input": "...", "parameters": {...}},
      {"model": "onnx://phi-3-mini", "input": "...", "parameters": {...}}
    ]
  }'
```

### Models

#### GET /v1/models

List available models.

```json
{
  "models": [
    {
      "model_id": "onnx://phi-3-mini",
      "provider": "onnx",
      "backend": "onnx_runtime",
      "context_length": 4096,
      "supported_parameters": ["temperature", "max_tokens", "top_p"],
      "is_available": true
    }
  ],
  "default_model": "phi-3-mini"
}
```

### Health

#### GET /health

Liveness probe.

```json
{
  "status": "healthy",
  "version": "0.1.0",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### GET /ready

Readiness probe (checks RunPod connectivity).

```json
{
  "status": "ready",
  "clusters": [...],
  "total_clusters": 3,
  "healthy_clusters": 3,
  "version": "0.1.0"
}
```

#### GET /metrics

Prometheus metrics.

```
# HELP ai_gateway_up AI Gateway service availability
# TYPE ai_gateway_up gauge
ai_gateway_up 1
...
```

## Model Identifiers

| Prefix | Provider | Backend |
|--------|----------|---------|
| `onnx://` | ONNX Runtime | Self-hosted ONNX model |
| `openai://` | OpenAI-compatible | OpenAI API |
| `runpod://` | RunPod | RunPod GPU cluster |

## Features

### Circuit Breaker

Automatically opens circuit when backend failure threshold is reached to prevent cascading failures.

```python
# Configuration
CIRCUIT_BREAKER_FAILURE_THRESHOLD = 5
CIRCUIT_BREAKER_RECOVERY_TIMEOUT_SECONDS = 60
```

### Rate Limiting

Per-tenant rate limiting using token bucket algorithm.

```python
# Configuration
RATE_LIMIT_REQUESTS = 100
RATE_LIMIT_WINDOW_SECONDS = 60
```

### Request Batching

Multiple requests are batched together for improved throughput.

```python
# Configuration
MAX_BATCH_SIZE = 8
BATCH_TIMEOUT_MS = 100
```

### Cost Tracking

Automatic cost calculation based on tokens and provider.

```python
# Configuration
ENABLE_COST_TRACKING = True
COST_PER_TOKEN = 0.00001
```

## Testing

```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=src --cov-report=html

# Run specific test file
pytest tests/test_api.py -v
```

## Development

```bash
# Install dev dependencies
pip install -e ai-gateway[dev]

# Run linter
ruff check src/

# Format code
ruff format src/

# Type check
mypy src/
```

## Integration with WASM Runtime

The AI Gateway is designed to be called from the WASM runtime via the `ai_infer` host function:

```rust
// Example WASM host function call
let request = json!({
    "model": "onnx://phi-3-mini",
    "input": base64_encode(input_data),
    "parameters": {"temperature": 0.7, "max_tokens": 100}
});

let response = host_call("ai_infer", serde_json::to_string(&request)?);
```

## License

MIT
