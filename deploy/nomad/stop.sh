#!/usr/bin/env bash
set -euo pipefail

NOMAD_ADDR="${NOMAD_ADDR:-http://127.0.0.1:4646}"

echo "=== Stopping FunctionFly Services ==="
echo ""

stop_job() {
    local job_name="$1"
    echo "Stopping $job_name..."

    if nomad job status "$job_name" &>/dev/null; then
        nomad job stop "$job_name" -detach
        echo "  Stopped $job_name"
    else
        echo "  $job_name not found, skipping"
    fi
}

echo "Stopping services in reverse order..."
stop_job "ai-gateway"
stop_job "agent-service"
stop_job "orchestrator-api"

echo ""
echo "=== All services stopped ==="