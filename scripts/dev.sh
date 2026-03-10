#!/bin/bash

# Fast development script - no Docker, direct local development

echo "🚀 Starting FunctionFly in development mode..."

# Load .env so JWT_SECRET etc. match publish script and generate_token
if [[ -f .env ]]; then
  set -a
  # shellcheck source=/dev/null
  source .env
  set +a
  echo "📋 Loaded .env (JWT_SECRET set for auth)"
fi

# Set environment variables for local development (no Docker)
# For Docker Postgres use DB_PORT=5434; for local Postgres use 5432 (default).
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-postgres}
export DB_NAME=${DB_NAME:-functionfly}
export DB_SSLMODE=${DB_SSLMODE:-disable}
export DEVELOPMENT=true
export REDIS_ADDR=${REDIS_ADDR:-localhost:6379}
export REDIS_PASSWORD=
export REDIS_DB=0

echo "📝 Environment variables set:"
echo "  DB_HOST=$DB_HOST"
echo "  DB_PORT=$DB_PORT"
echo "  DB_USER=$DB_USER"
echo "  DB_NAME=$DB_NAME"

# Check if PostgreSQL is running
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER >/dev/null 2>&1; then
    echo "❌ PostgreSQL is not running on $DB_HOST:$DB_PORT"
    echo "💡 Start PostgreSQL with: sudo systemctl start postgresql"
    echo "💡 Or install PostgreSQL if not installed"
    exit 1
fi

echo "✅ PostgreSQL is running"

# Check if database exists
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    echo "📦 Creating database '$DB_NAME'..."
    PGPASSWORD=$DB_PASSWORD createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME
fi

echo "✅ Database '$DB_NAME' exists"

# Start the API server
echo "🔥 Starting API server..."
echo "📡 API will be available at: http://localhost:8080"
echo "🔄 Hot reload enabled - changes will restart automatically"

# Use air for hot reloading if available, otherwise use go run
if command -v air >/dev/null 2>&1; then
    echo "✨ Using air for hot reloading"
    air
else
    echo "⚠️  air not found. Install with: go install github.com/cosmtrek/air@latest"
    echo "🔄 Using go run (manual restart required)"
    go run ./cmd/orchestrator-api
fi
