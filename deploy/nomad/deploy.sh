#!/usr/bin/env bash
set -euo pipefail

NOMAD_ADDR="${NOMAD_ADDR:-http://127.0.0.1:4646}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== FunctionFly Nomad Deployment ==="
echo "Nomad Address: $NOMAD_ADDR"
echo ""

wait_for_nomad() {
    echo "Waiting for Nomad to be available..."
    local max_attempts=30
    local attempt=1
    while ! nomad job status &>/dev/null; do
        if (( attempt >= max_attempts )); then
            echo "Error: Nomad not available after $max_attempts attempts"
            exit 1
        fi
        echo "  Attempt $attempt/$max_attempts..."
        sleep 2
        ((attempt++))
    done
    echo "Nomad is available"
}

deploy_job() {
    local job_file="$1"
    local job_name
    job_name=$(basename "$job_file" .nomad)

    echo ""
    echo "--- Deploying $job_name ---"
    nomad job run "$job_file"
    echo "Deployed $job_name"
}

wait_for_healthy() {
    local job_name="$1"
    local timeout="${2:-300}"

    echo "Waiting for $job_name to become healthy..."
    local start_time=$(date +%s)
    local healthy=false

    while true; do
        local status
        status=$(nomad job status "$job_name" 2>/dev/null | grep "Status" | awk '{print $2}' || echo "unknown")

        if [[ "$status" == "running" ]]; then
            local task_status
            task_status=$(nomad job status "$job_name" 2>/dev/null | grep -A 20 "Task Group" | grep "Status" | head -1 | awk '{print $2}' || echo "unknown")
            if [[ "$task_status" == "running" ]]; then
                healthy=true
                break
            fi
        fi

        local elapsed=$(($(date +%s) - start_time))
        if (( elapsed >= timeout )); then
            echo "Timeout waiting for $job_name to become healthy"
            return 1
        fi

        echo "  Status: $status, waiting..."
        sleep 5
    done

    if $healthy; then
        echo "$job_name is healthy"
        return 0
    else
        echo "$job_name failed to become healthy"
        return 1
    fi
}

main() {
    wait_for_nomad

    echo ""
    echo "Deploying jobs in order..."
    echo "1. orchestrator-api"
    echo "2. agent-service"
    echo "3. ai-gateway"

    deploy_job "$SCRIPT_DIR/jobs/orchestrator-api.nomad"
    wait_for_healthy "orchestrator-api" 120

    deploy_job "$SCRIPT_DIR/jobs/agent-service.nomad"
    wait_for_healthy "agent-service" 120

    deploy_job "$SCRIPT_DIR/jobs/ai-gateway.nomad"
    wait_for_healthy "ai-gateway" 300

    echo ""
    echo "=== Deployment Complete ==="
    echo ""
    echo "Job Status:"
    nomad job status | grep -E "(ID|Type|Status)" | head -20
}

main "$@"