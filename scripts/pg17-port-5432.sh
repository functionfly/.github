#!/usr/bin/env bash
# Configure PostgreSQL 17 to use port 5432 only (Debian/Ubuntu).
# Run: bash scripts/pg17-port-5432.sh   (will prompt for sudo)
# See: docs/LOCAL_POSTGRES_17.md section 7
set -e

if [ "$(id -u)" -ne 0 ]; then
  exec sudo bash "$0" "$@"
fi

CONF="${CONF:-/etc/postgresql/17/main/postgresql.conf}"
if [ ! -f "$CONF" ]; then
  echo "Config not found: $CONF"
  echo "Set CONF=/path/to/postgresql.conf if your PG 17 config is elsewhere."
  exit 1
fi

echo "Configuring PG 17 to use port 5432: $CONF"
cp -a "$CONF" "${CONF}.bak.$(date +%Y%m%d%H%M%S)"
sed -i "s/^#*port = .*/port = 5432/" "$CONF"
echo "Done. Restart PG 17: sudo pg_ctlcluster 17 main restart"
echo "If PG 16 was on 5432, stop it first: sudo pg_ctlcluster 16 main stop"
