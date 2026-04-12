#!/usr/bin/env bash
# Start FlyMind (ai-service) if nothing responds on AI_SERVICE_PORT, then run the orchestrator.
# When this script starts ai-service, it stops FlyMind when the orchestrator exits (trap).
#
# Example:
#   set -a && source .env && set +a
#   unset DEVELOPMENT
#   export REDIS_ADDR=localhost:6379 SKIP_MIGRATION_VALIDATION=true VERIFICATION_ENABLED=false
#   ./scripts/run-orchestrator-with-ai.sh ./bin/orchestrator-api --skip-migrations
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
AI_PORT="${AI_SERVICE_PORT:-18081}"
HEALTH_URL="${AI_SERVICE_HEALTH_URL:-http://127.0.0.1:${AI_PORT}/health}"
# Ensure the orchestrator points to the same FlyMind endpoint this wrapper manages.
export AI_SERVICE_URL="${AI_SERVICE_URL:-http://127.0.0.1:${AI_PORT}}"

ai_up() {
  if command -v curl >/dev/null 2>&1; then
    curl -sf "$HEALTH_URL" >/dev/null 2>&1
  else
    (echo >/dev/tcp/127.0.0.1/"$AI_PORT") >/dev/null 2>&1
  fi
}

AI_PID=""
if ai_up; then
  echo "FlyMind ai-service already running (${HEALTH_URL})."
elif command -v uv >/dev/null 2>&1 && [ -f "$ROOT/ai-service/pyproject.toml" ]; then
  echo "Starting FlyMind ai-service on 127.0.0.1:${AI_PORT} ..."
  (
    cd "$ROOT/ai-service"
    unset VIRTUAL_ENV
    set -a
    [ -f .env ] && . ./.env
    set +a
    # Disable RAG to prevent Ollama blocking during startup
    export ENABLE_RAG="${ENABLE_RAG:-false}"
    export RAG_EMBEDDING_PROVIDER="${RAG_EMBEDDING_PROVIDER:-openai}"
    # Force local Redis for ai-service dev mode (ignore Upstash URL from root .env)
    export REDIS_URL="redis://localhost:6379"
    exec env PYTHONPATH=. uv run uvicorn src.main:app --host 127.0.0.1 --port "$AI_PORT"
  ) &
  AI_PID=$!
  for _ in $(seq 1 80); do
    if ai_up; then
      echo "FlyMind ai-service ready (${HEALTH_URL})."
      break
    fi
    sleep 0.25
  done
  if ! ai_up; then
    echo "WARN: FlyMind did not become healthy in time; orchestrator will use rule-based fallback until ai-service responds." >&2
  fi
else
  echo "WARN: uv not found or ai-service missing; install uv (https://docs.astral.sh/uv/) and run: cd ai-service && uv sync" >&2
fi

cleanup() {
  if [ -n "${AI_PID:-}" ]; then
    kill "$AI_PID" 2>/dev/null || true
    wait "$AI_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

# Do not use exec when we started FlyMind — otherwise EXIT trap never runs and the child keeps running orphaned.
if [ -n "$AI_PID" ]; then
  "$@"
  STATUS=$?
  cleanup
  trap - EXIT INT TERM
  exit "$STATUS"
fi

exec "$@"
