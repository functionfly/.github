#!/bin/bash
set -euo pipefail

EXECUTION_ID="${FLY_EXECUTION_ID:-}"
ORCHESTRATOR_URL="${FUNCTIONFLY_ORCHESTRATOR_URL:-http://localhost:8080}"
SECRET="${FUNCTIONFLY_MACHINE_SECRET:-}"
TIMEOUT_SECONDS="${FLY_TIMEOUT_SECONDS:-30}"

if [[ -z "$EXECUTION_ID" ]]; then
    echo "ERROR: FLY_EXECUTION_ID is required"
    exit 1
fi

if [[ -z "$ORCHESTRATOR_URL" ]]; then
    echo "ERROR: FUNCTIONFLY_ORCHESTRATOR_URL is required"
    exit 1
fi

echo "Fetching code for execution: $EXECUTION_ID"

AUTH_HEADER=""
if [[ -n "$SECRET" ]]; then
    AUTH_HEADER="-H X-Machine-Secret: ${SECRET}"
fi

RESPONSE=$(curl -sSf \
    -H "Content-Type: application/json" \
    ${AUTH_HEADER} \
    "${ORCHESTRATOR_URL}/api/runtime/flymachines/code/${EXECUTION_ID}")

if [[ $? -ne 0 ]]; then
    echo "ERROR: Failed to fetch code from orchestrator"
    exit 1
fi

SOURCE=$(echo "$RESPONSE" | jq -r '.source // empty')
INPUT=$(echo "$RESPONSE" | jq -r '.input // "{}"')

if [[ -z "$SOURCE" ]]; then
    echo "ERROR: No source code in response"
    exit 1
fi

echo "Executing Python code..."

START_TIME=$(date +%s)

RESULT=$(python3.12 -c "$SOURCE" 2>&1 <<< "$INPUT")
EXIT_CODE=$?

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo "Execution completed in ${DURATION}s with exit code ${EXIT_CODE}"

if [[ -n "$SECRET" ]]; then
    RESULT_PAYLOAD=$(jq -n \
        --arg result "$RESULT" \
        --argjson exit_code $EXIT_CODE \
        --argjson duration $DURATION \
        '{
            result: $result,
            exit_code: $exit_code,
            duration_seconds: $duration,
            executed_at: (now | strftime("%Y-%m-%dT%H:%M:%SZ"))
        }')

    curl -sSf -X POST \
        -H "Content-Type: application/json" \
        -H "X-Machine-Secret: ${SECRET}" \
        -d "$RESULT_PAYLOAD" \
        "${ORCHESTRATOR_URL}/api/runtime/flymachines/result/${EXECUTION_ID}" || true
fi

exit $EXIT_CODE
