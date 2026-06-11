#!/bin/bash
# Usage: ./scripts/run-rq-worker.sh
#
# Starts the RQ worker for audit log processing
set -e

cd "$(dirname "$0")/.."

export PYTHONPATH="${PYTHONPATH}:$(pwd)/src"

echo "Starting RQ worker for audit queue..."
exec uv run rq worker audit --url "redis://${REDIS_ADDR:-localhost:6379}/0"
