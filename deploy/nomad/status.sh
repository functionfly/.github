#!/usr/bin/env bash
set -euo pipefail

NOMAD_ADDR="${NOMAD_ADDR:-http://127.0.0.1:4646}"

echo "=== FunctionFly Nomad Status ==="
echo ""

jobs=("orchestrator-api" "agent-service" "ai-gateway")

for job in "${jobs[@]}"; do
    echo "--- $job ---"
    if nomad job status "$job" &>/dev/null; then
        nomad job status "$job" 2>/dev/null | grep -E "(ID|Type|Status|Submit Time|Task Group|Instances)" | head -15
    else
        echo "  Not deployed"
    fi
    echo ""
done

echo "--- Node Status ---"
nomad node status 2>/dev/null | grep -E "(ID|DC|Status|Node Pool|GPUs)" | head -20 || echo "No nodes available"