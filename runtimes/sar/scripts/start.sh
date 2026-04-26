#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SAR_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== FunctionFly SAR (Stateful Agent Runtime) ==="

# Check NATS
if ! pgrep -x nats-server > /dev/null 2>&1; then
    echo "Starting NATS server on port 4222..."
    nohup nats-server -p 4222 > /tmp/nats.log 2>&1 &
    sleep 1
    if pgrep -x nats-server > /dev/null 2>&1; then
        echo "  NATS started (PID: $(pgrep -x nats-server))"
    else
        echo "  WARNING: NATS failed to start. Check /tmp/nats.log"
    fi
else
    echo "  NATS already running (PID: $(pgrep -x nats-server))"
fi

# Build if needed
if [ ! -f "$SAR_DIR/target/debug/functionfly-sar" ]; then
    echo "Building SAR..."
    cargo build --manifest-path "$SAR_DIR/Cargo.toml"
fi

# Start SAR
echo "Starting SAR on port ${PORT:-8082}..."
exec "$SAR_DIR/target/debug/functionfly-sar" "$@"
