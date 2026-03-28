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

From the `ai-service` directory (uses `ai-service/.env`; clear a conflicting repo-root `VIRTUAL_ENV` if `uv` warns):

```bash
unset VIRTUAL_ENV
uv sync
PYTHONPATH=. uv run uvicorn src.main:app --host 127.0.0.1 --port 18081
```

Or use the repo helper (from repo root):

```bash
./scripts/run-ai-service.sh
```

## Configuration

See `.env.example`. Key settings: `DATABASE_URL`, `REDIS_URL`, provider API keys.
