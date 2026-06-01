#!/bin/bash
# bore Tunnel – expose local services via bore
# Usage: ./deploy/bore-tunnel.example.sh
#
# Prerequisites:
# 1. Install: cargo install bore-cli
# 2. Or: brew install bore-cli
#
# This script exposes both the orchestrator API (port 8080) and dashboard (port 3000)
# via bore tunnels. Press Ctrl+C to stop.

set -e

BORE="${BORE:-$HOME/.cargo/bin/bore}"
API_PORT=${API_PORT:-8080}
DASHBOARD_PORT=${DASHBOARD_PORT:-3000}

if ! command -v bore &>/dev/null && [ ! -x "$BORE" ]; then
  echo "bore not found. Install: cargo install bore-cli"
  exit 1
fi

echo "Starting bore tunnels..."
echo "  API:       http://localhost:$API_PORT"
echo "  Dashboard: http://localhost:$DASHBOARD_PORT"
echo
echo "Press Ctrl+C to stop."
echo

$BORE local "$API_PORT" --to bore.pub &
BORE_API_PID=$!

$BORE local "$DASHBOARD_PORT" --to bore.pub &
BORE_DASH_PID=$!

wait $BORE_API_PID $BORE_DASH_PID