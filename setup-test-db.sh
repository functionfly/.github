# Test Database Setup Script
# Start PostgreSQL container
docker compose up -d postgres

# Wait for database to be ready
sleep 5

# Create test database
docker compose exec postgres psql -U postgres -c "CREATE DATABASE functionfly_test;"

# Run migrations on test database
DB_HOST=localhost DB_PORT=5434 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=functionfly_test DB_SSLMODE=disable go run ./cmd/migrate run

# Verify tables were created
docker compose exec postgres psql -U postgres -d functionfly_test -c "\dt"

# Run tests
go test -v ./internal/storage -run TestPostgresTestSuite/TestNewPostgresDB
