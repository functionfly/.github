# Production-ready Python libraries (ai-service)

Optional dependency groups in `pyproject.toml` and how to use them.

## Install optional groups

```bash
# All production extras (observability + resilience + errors)
pip install .[prod]
# or: uv sync --extra prod

# Or individually
pip install .[observability] .[resilience] .[errors]
```

---

## 1. Observability (`[observability]`)

| Library | Purpose |
|--------|--------|
| **OpenTelemetry** (api, sdk, instrumentation-fastapi/httpx/redis, exporter-otlp) | Distributed tracing and metrics; wire requests across FastAPI, httpx, and Redis. Export to Jaeger, Tempo, or any OTLP backend. |
| **structlog** | Structured logging with key-value fields; integrates with your existing `LogContext` (tenant_id, request_id, etc.) and works alongside OpenTelemetry trace IDs. |

**Quick wins**

- Add OTLP exporter and set `OTEL_EXPORTER_OTLP_ENDPOINT` in production to get traces.
- Replace or wrap `logging` in hot paths with `structlog.get_logger()` and bind `request_id`/`tenant_id` per request (e.g. middleware or contextvars).

---

## 2. Resilience (`[resilience]`)

| Library | Purpose |
|--------|--------|
| **tenacity** | Retries with exponential backoff, jitter, and stop conditions. Use for orchestrator HTTP calls, LLM provider calls, and Redis where you already have `RetryConfig` but want consistent behavior and retry-on-exception only. |

**Quick wins**

- Wrap `OrchestratorClient` (e.g. `trigger_prewarm`, `get_function`) with `@retry(stop=stop_after_attempt(3), wait=wait_exponential(min=1, max=30))`.
- Use the same pattern for any `httpx` call that should be retried on 5xx or connection errors.

---

## 3. Error tracking (`[errors]`)

| Library | Purpose |
|--------|--------|
| **sentry-sdk[fastapi]** | Capture unhandled exceptions and optional breadcrumbs; FastAPI integration adds request context and user/tenant scope. |

**Quick wins**

- Set `SENTRY_DSN` in production; init Sentry early in `main.py` and use `sentry_sdk.set_tag("tenant_id", ...)` in middleware so all errors are tagged.

---

## 4. Already in the stack (no new libs)

- **Retries**: You have `RetryConfig` and backoff in `BaseProvider`; use **tenacity** where you want the same behavior for non-LLM calls (orchestrator, Redis, etc.).
- **Rate limiting**: Your in-process `RateLimiter` is fine for single-instance; for multi-instance consider Redis-based limits or API gateway limits later.
- **Config**: `pydantic-settings` with `.env` is production-suitable; keep secrets out of env in prod (e.g. vault or secret manager).
- **Health**: Your health checker + dependency checks (Redis, etc.) are good; add orchestrator and DB to health if they are critical.

---

## 5. Optional future additions

- **arq** or **celery**: If background jobs (e.g. `workers/tasks.py`) need persistence, retries, and dead-letter queues beyond the current in-process scheduler.
- **slowapi**: API-level rate limiting (e.g. per-IP or per-tenant) if you move rate limits to the HTTP layer.
- **cryptography**: If you ever need to hash or encrypt secrets beyond what you already do for the vault.

Install only what you need; start with `[observability]` and `[errors]` for production, then add `[resilience]` where you standardize retries.
