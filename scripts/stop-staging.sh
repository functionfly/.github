#!/bin/bash

# FunctionFly Staging Environment Stopper
# This script stops the staging environment

echo "🛑 Stopping FunctionFly Staging Environment..."

# Stop staging containers
docker-compose -f docker-compose.staging.yml down

echo "✅ Staging environment stopped successfully!"

# Optional: Remove volumes (uncomment if you want to clean up data)
# echo "🧹 Removing staging volumes..."
# docker volume rm functionfly_redis_staging_data functionfly_caddy_staging_data || true

echo ""
echo "💡 To restart: ./scripts/run-staging.sh"
echo "💡 To view logs: docker-compose -f docker-compose.staging.yml logs"