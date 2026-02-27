#!/bin/bash

# Setup script for PostgreSQL read replica
# This script configures streaming replication from primary to replica

set -euo pipefail

echo "Setting up PostgreSQL read replica..."

# Wait for primary to be ready
echo "Waiting for primary database to be ready..."
timeout=300
counter=0
while ! docker-compose -f docker-compose.production.yml exec -T postgres pg_isready -U functionfly_prod -h localhost -p 5432 >/dev/null 2>&1; do
    if [ $counter -gt $timeout ]; then
        echo "Timeout waiting for primary database"
        exit 1
    fi
    counter=$((counter + 5))
    echo "Waiting... ($counter/$timeout seconds)"
    sleep 5
done

echo "Primary database is ready"

# Wait for replica to be ready
echo "Waiting for replica database to be ready..."
counter=0
while ! docker-compose -f docker-compose.production.yml exec -T postgres-replica pg_isready -U functionfly_prod -h localhost -p 5432 >/dev/null 2>&1; do
    if [ $counter -gt $timeout ]; then
        echo "Timeout waiting for replica database"
        exit 1
    fi
    counter=$((counter + 5))
    echo "Waiting... ($counter/$timeout seconds)"
    sleep 5
done

echo "Replica database is ready"

# Create replication user on primary (if not exists)
echo "Ensuring replication user exists on primary..."
docker-compose -f docker-compose.production.yml exec -T postgres psql -U functionfly_prod -d functionfly_prod -c "
DO \$\$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'replicator') THEN
      CREATE ROLE replicator LOGIN REPLICATION PASSWORD '${DB_PASSWORD}';
   END IF;
END
\$\$;"

echo "Read replica setup completed successfully!"
echo ""
echo "Replica Status:"
echo "- Primary: localhost:5432"
echo "- Replica: localhost:5433"
echo ""
echo "To check replication status:"
echo "docker-compose -f docker-compose.production.yml exec postgres psql -U functionfly_prod -d functionfly_prod -c \"SELECT * FROM pg_stat_replication;\""
echo ""
echo "To check replica lag:"
echo "docker-compose -f docker-compose.production.yml exec postgres-replica psql -U functionfly_prod -d functionfly_prod -c \"SELECT now() - pg_last_xact_replay_timestamp() AS replication_lag;\""