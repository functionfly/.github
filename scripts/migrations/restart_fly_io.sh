#!/bin/bash
# restart_fly_io.sh - Restart Fly.io app to clear caches after tenant migration
# This clears prepared statements and in-memory caches after the database migration

set -e

FLY_APP="${FLY_APP:-functionfly-control}"

echo "=== Restarting Fly.io app: $FLY_APP ==="
echo "This will clear prepared statements and in-memory caches after the migration"
echo ""

# Method 1: Rolling restart (zero downtime)
echo "Option 1: Rolling restart (recommended for production)"
echo "  fly apps restart $FLY_APP"
echo ""

# Method 2: Deploy new version (if code changes exist)
echo "Option 2: Deploy (if you have code changes to deploy)"
echo "  fly deploy --app $FLY_APP"
echo ""

# Method 3: Restart specific machines
echo "Option 3: Restart specific machines"
echo "  fly machines list --app $FLY_APP"
echo "  fly machines restart <machine-id> --app $FLY_APP"
echo ""

# Check current status
echo "=== Current Fly.io Status ==="
fly status --app "$FLY_APP" 2>/dev/null || echo "Fly CLI not configured or app not found"

echo ""
echo "=== Redis Cache (if using Redis) ==="
echo "To clear Redis cache (if you suspect stale cached data):"
echo "  fly redis connect --app $FLY_APP"
echo "  (then run: FLUSHDB)"
echo ""
echo "Or if using Upstash/External Redis, connect via their console and clear keys matching 'user:*'"

echo ""
echo "=== After Restart ==="
echo "1. Wait for health checks to pass: fly status --app $FLY_APP"
echo "2. Check logs: fly logs --app $FLY_APP"
echo "3. Test login with affected user"
