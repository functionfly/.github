#!/bin/bash

# FunctionFly Staging Environment Runner
# This script starts the staging environment using Docker Compose

set -e

echo "🚀 Starting FunctionFly Staging Environment..."

# Check if .env.staging exists
if [ ! -f ".env.staging" ]; then
    echo "❌ Error: .env.staging file not found!"
    echo "   Please create .env.staging from .env.example"
    exit 1
fi

# Load staging environment variables
export $(grep -v '^#' .env.staging | xargs)

# Stop any existing staging containers
echo "🛑 Stopping existing staging containers..."
docker-compose -f docker-compose.staging.yml down || true

# Start staging environment
echo "🏗️  Building and starting staging environment..."
docker-compose -f docker-compose.staging.yml up --build -d

echo "✅ Staging environment started successfully!"
echo ""
echo "📊 Services:"
echo "   • Orchestrator API: http://localhost:8082"
echo "   • Caddy Proxy: http://localhost:8083"
echo "   • Health Monitor: Running in background"
echo "   • Redis: localhost:6380"
echo ""
echo "🩺 Health Check: curl http://localhost:8082/health"
echo ""
echo "🛑 To stop: docker-compose -f docker-compose.staging.yml down"
echo "📝 Logs: docker-compose -f docker-compose.staging.yml logs -f"