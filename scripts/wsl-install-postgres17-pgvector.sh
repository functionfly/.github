#!/usr/bin/env bash
# Install PostgreSQL 17 with pgvector on WSL (Ubuntu/Debian).
#
# Fresh install: run as-is. It adds the PGDG repo, installs postgresql-17 and
# pgvector (package or from source), and optionally creates DB 'functionfly'.
#
# If you already have Postgres 15/16: this installs 17 alongside. To switch to 17
# as default and migrate data: pg_dropcluster, pg_upgradecluster, or use 17 on
# a different port and point DB_PORT to it.
#
# Run: bash scripts/wsl-install-postgres17-pgvector.sh
# Or:  chmod +x scripts/wsl-install-postgres17-pgvector.sh && ./scripts/wsl-install-postgres17-pgvector.sh
set -e

# Detect WSL (optional, for messaging)
if grep -qi microsoft /proc/version 2>/dev/null; then
  echo "WSL detected."
fi

# Require Ubuntu/Debian
if [ ! -f /etc/debian_version ]; then
  echo "This script targets Debian/Ubuntu. /etc/debian_version not found."
  exit 1
fi

echo "=== Installing PostgreSQL 17 and pgvector on WSL ==="

# 1) PGDG repository (official PostgreSQL packages)
echo "Adding PostgreSQL APT repository (PGDG)..."
sudo apt-get update -qq
sudo apt-get install -y ca-certificates curl

# Use PGDG setup script if present, else add repo manually
if [ -x /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh ]; then
  sudo /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh
else
  sudo install -d /usr/share/postgresql-common/pgdg
  sudo curl -sSf -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc \
    https://www.postgresql.org/media/keys/ACCC4CF8.asc
  # codename for -pgdg (e.g. jammy-pgdg, noble-pgdg)
  CODENAME=$(lsb_release -cs 2>/dev/null || echo "bookworm")
  echo "deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt ${CODENAME}-pgdg main" \
    | sudo tee /etc/apt/sources.list.d/pgdg.list > /dev/null
fi

sudo apt-get update -qq

# 2) Install PostgreSQL 17
echo "Installing PostgreSQL 17..."
sudo apt-get install -y postgresql-17 postgresql-client-17

# 3) Install pgvector for PostgreSQL 17
# Package name may be postgresql-17-pgvector (PGDG/Ubuntu) or we build from source
if apt-cache show postgresql-17-pgvector &>/dev/null; then
  echo "Installing postgresql-17-pgvector from package..."
  sudo apt-get install -y postgresql-17-pgvector
else
  echo "postgresql-17-pgvector package not found. Building pgvector from source..."
  sudo apt-get install -y postgresql-server-dev-17 git build-essential
  BUILD_DIR="${TMPDIR:-/tmp}/pgvector-build"
  rm -rf "$BUILD_DIR"
  git clone --depth 1 --branch v0.8.0 https://github.com/pgvector/pgvector.git "$BUILD_DIR"
  cd "$BUILD_DIR"
  make
  sudo make install
  cd -
  rm -rf "$BUILD_DIR"
fi

# 4) Start and enable PostgreSQL 17
echo "Starting PostgreSQL 17..."
sudo service postgresql start 2>/dev/null || sudo systemctl start postgresql@17 2>/dev/null || true
sudo service postgresql status 2>/dev/null || sudo systemctl status postgresql@17 --no-pager 2>/dev/null || true

# 5) Optional: create functionfly DB and enable extensions (skip if you already have DB)
echo ""
echo "PostgreSQL 17 and pgvector are installed."
echo ""
REPLY=""
if [ -t 0 ]; then
  read -r -p "Create database 'functionfly' and user (password: postgres) and enable pgvector? [y/N] " REPLY
fi
if [[ "$REPLY" =~ ^[yY]$ ]]; then
  sudo -u postgres psql -c "CREATE USER $USER WITH SUPERUSER CREATEDB PASSWORD 'postgres';" 2>/dev/null || true
  sudo -u postgres psql -c "CREATE DATABASE functionfly OWNER $USER;" 2>/dev/null || true
  sudo -u postgres psql -d functionfly -c "CREATE EXTENSION IF NOT EXISTS vector; CREATE EXTENSION IF NOT EXISTS pg_trgm;" 2>/dev/null || true
  echo "Database 'functionfly' ready. Run: make migrate-local && make setup"
else
  echo "Manual steps:"
  echo "  sudo -u postgres psql -c \"CREATE USER $USER WITH SUPERUSER CREATEDB PASSWORD 'postgres';\""
  echo "  sudo -u postgres psql -c \"CREATE DATABASE functionfly OWNER $USER;\""
  echo "  sudo -u postgres psql -d functionfly -c 'CREATE EXTENSION IF NOT EXISTS vector;'"
  echo "  make migrate-local && make setup"
fi
echo ""
echo "Version: $(sudo -u postgres psql -t -c 'SELECT version();' 2>/dev/null | head -1)"
