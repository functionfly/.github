#!/usr/bin/env bash
# Start FlyMind (ai-service) on localhost. Use from repo root or via run-orchestrator-with-ai.sh.
# Clears parent VIRTUAL_ENV so uv uses ai-service/.venv (avoids uv warning when repo .venv is active).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/ai-service"

PORT="${AI_SERVICE_PORT:-18081}"
HOST="${AI_SERVICE_HOST:-127.0.0.1}"

if ! command -v uv >/dev/null 2>&1; then
  echo "ERROR: uv is required. Install: https://docs.astral.sh/uv/" >&2
  exit 1
fi

set -a
[ -f .env ] && . ./.env
set +a

unset VIRTUAL_ENV

echo "FlyMind ai-service: http://${HOST}:${PORT}/health (docs: /docs)"
# PYTHONPATH=. + src.main so relative imports in main.py resolve (main:app --app-dir src breaks package context)
exec env PYTHONPATH=. uv run uvicorn src.main:app --host "$HOST" --port "$PORT"
