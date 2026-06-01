#!/bin/bash
# ngrok Tunnel – expose the orchestrator API and dashboard locally via ngrok
# Usage: ./deploy/ngrok-tunnel.example.sh
#
# Prerequisites:
# 1. Install ngrok: https://ngrok.com/download (or `brew install ngrok`)
# 2. Sign up at https://ngrok.com and connect your authtoken
# 3. Run: ngrok config add-authtoken <YOUR_AUTH_TOKEN>
#
# This script exposes both the orchestrator API (port 8080) and dashboard (port 3000)
# via ngrok. Press Ctrl+C to stop.

API_PORT=${API_PORT:-8080}
DASHBOARD_PORT=${DASHBOARD_PORT:-3000}

if ! command -v ngrok &>/dev/null; then
  echo "ngrok not found. Install from: https://ngrok.com/download"
  exit 1
fi

echo "Starting ngrok tunnels..."
echo "  API:       http://localhost:$API_PORT"
echo "  Dashboard: http://localhost:$DASHBOARD_PORT"
echo
echo "Press Ctrl+C to stop."
echo

ngrok http "$API_PORT" --log=stdout &
NGROK_PID=$!
sleep 3

ngrok http "$DASHBOARD_PORT" --log=stdout &
sleep 2

echo
echo "Tunnel URLs:"
curl -s http://localhost:4040/api/tunnels 2>/dev/null | python3 -c "
import json,sys
data=json.load(sys.stdin)
for t in data.get('tunnels',[]):
    print(f\"  {t.get('proto','?')}://{t.get('public_url','').replace('https://','')}\")
" 2>/dev/null || echo "  (check ngrok UI at http://localhost:4040)"

wait $NGROK_PID
