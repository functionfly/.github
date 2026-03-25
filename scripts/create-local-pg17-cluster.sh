#!/usr/bin/env bash
# Create a local PostgreSQL 17 cluster on port 5432 with pgvector and DB 'functionfly'.
# Run from repo root: bash scripts/create-local-pg17-cluster.sh
# See: docs/LOCAL_POSTGRES_17.md
set -e

if [ ! -f /etc/debian_version ]; then
  echo "This script targets Debian/Ubuntu. For other distros or Docker, see docs/LOCAL_POSTGRES_17.md section 8."
  exit 1
fi

echo "=== 1) Add PGDG repository ==="
# Remove ALL PGDG source files (both .list and .sources/deb822 formats) plus cached lists.
# postgresql-common may have written pgdg.sources; remove it so the first update is PGDG-free.
sudo rm -f \
  /etc/apt/sources.list.d/pgdg.list \
  /etc/apt/sources.list.d/pgdg.sources \
  /etc/apt/sources.list.d/postgresql.list \
  /var/lib/apt/lists/apt.postgresql.org* \
  2>/dev/null || true
# Update and install prerequisites from Ubuntu repos only (no PGDG yet).
sudo apt-get update -qq
sudo apt-get install -y ca-certificates curl lsb-release

# Add PGDG signing key and source manually.
sudo install -d /usr/share/postgresql-common/pgdg
sudo curl -sSfL -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc \
  https://www.postgresql.org/media/keys/ACCC4CF8.asc
CODENAME=$(lsb_release -cs 2>/dev/null || echo "noble")
echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt ${CODENAME}-pgdg main" \
  | sudo tee /etc/apt/sources.list.d/pgdg.list > /dev/null

echo "=== 2) Update PGDG index (forcing .gz to bypass .bz2 mirror sync mismatch) ==="
# The CDN inconsistency is specific to Packages.bz2; .gz is served consistently.
sudo rm -f /var/lib/apt/lists/apt.postgresql.org* 2>/dev/null || true
sudo apt-get update -o Acquire::CompressionTypes::Order::=gz

echo "=== 3) Install PostgreSQL 17 and pgvector ==="
sudo apt-get install -y postgresql-17 postgresql-client-17 postgresql-common
if apt-cache show postgresql-17-pgvector &>/dev/null; then
  sudo apt-get install -y postgresql-17-pgvector
else
  sudo apt-get install -y postgresql-17-contrib 2>/dev/null || true
fi
if apt-cache show postgresql-17-pgvector &>/dev/null; then
  sudo apt-get install -y postgresql-17-pgvector
else
  echo "Installing postgresql-17-contrib (pgvector may need to be built from source; see docs/LOCAL_POSTGRES_17.md)"
  sudo apt-get install -y postgresql-17-contrib 2>/dev/null || true
fi

echo "=== 4) Create and start cluster '17 main' ==="
# pg_createcluster is idempotent via the OR fallback; if config exists, just start it.
sudo pg_createcluster 17 main 2>/dev/null || true
sudo pg_ctlcluster 17 main start 2>/dev/null || true

# Ensure port 5432
PGPORT=$(sudo pg_lsclusters | awk '/17 main/ {print $3}')
if [ -n "$PGPORT" ] && [ "$PGPORT" != "5432" ]; then
  echo "=== 5) Configure PG 17 to use port 5432 ==="
  SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
  bash "$SCRIPT_DIR/pg17-port-5432.sh"
  sudo pg_ctlcluster 17 main restart
fi

echo "=== 6) Create database 'functionfly' and set postgres password ==="
sudo -u postgres psql -p 5432 -c "ALTER USER postgres PASSWORD 'postgres';" 2>/dev/null || true
sudo -u postgres psql -p 5432 -c "CREATE DATABASE functionfly;" || true

echo "=== 7) Enable extensions ==="
sudo -u postgres psql -p 5432 -d functionfly -v ON_ERROR_STOP=0 -c "
  CREATE EXTENSION IF NOT EXISTS \"pg_trgm\";
  CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";
  CREATE EXTENSION IF NOT EXISTS vector;
  CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";
" || true

echo ""
echo "Local PostgreSQL 17 cluster is ready on port 5432."
echo "  DB: functionfly   User: postgres   Password: postgres"
echo ""
echo "Next: make migrate-local && make setup-local"
echo "  Or:  ./scripts/run-create-admin.sh -email admin@functionfly.local -password admin123 -role super_admin"
echo ""
