#!/bin/bash
# Graceful shutdown wrapper for orchestrator-api
# Handles SIGTERM to allow graceful shutdown before killing

APP_NAME="orchestrator-api"
SHUTDOWN_TIMEOUT=30

# Get the PID of the main process
PID_FILE="/tmp/orchestrator-api.pid"

# Trap SIGTERM
trap 'handle_sigterm' SIGTERM

handle_sigterm() {
    echo "[$APP_NAME] Received SIGTERM, initiating graceful shutdown..."

    # If we have a PID file, send SIGTERM to the actual process
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            # Send SIGTERM to the process
            kill -TERM "$PID"

            # Wait for the process to exit
            WAITED=0
            while kill -0 "$PID" 2>/dev/null && [ $WAITED -lt $((SHUTDOWN_TIMEOUT * 10)) ]; do
                sleep 0.1
                WAITED=$((WAITED + 1))
            done

            # If still running, force kill
            if kill -0 "$PID" 2>/dev/null; then
                echo "[$APP_NAME] Process did not exit gracefully, forcing..."
                kill -9 "$PID" 2>/dev/null
            fi
        fi
        rm -f "$PID_FILE"
    fi

    echo "[$APP_NAME] Graceful shutdown complete"
    exit 0
}

# Start the main process in background and save PID
./orchestrator-api "$@" &
PID=$!
echo $PID > "$PID_FILE"

# Wait for the process
wait $PID
EXIT_CODE=$?

# Clean up PID file on exit
rm -f "$PID_FILE"

exit $EXIT_CODE
