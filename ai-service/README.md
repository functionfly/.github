# FlyMind AI Service

Intelligent capabilities for FunctionFly: LLM providers (OpenAI, Anthropic, Ollama), embeddings, caching, and health checks.

## Setup

```bash
uv sync
```

Optional: self-hosted content moderation (Detoxify) for toxicity/hate/violence:

```bash
uv sync --extra moderation
```

gRPC server (optional): generate stubs using the project venv (do not use system `pip`/`python`):

```bash
uv sync
uv run python scripts/generate_grpc.py
# then start the app; gRPC listens on 0.0.0.0:50051 by default
```

## Run

```bash
uv run uvicorn api.app:app --host 0.0.0.0 --port 8081
```

## Configuration

See `.env.example`. Key settings: `DATABASE_URL`, `REDIS_URL`, provider API keys.
