#!/usr/bin/env bash
# Fix apt issues blocking PostgreSQL 17 install:
# 1. PGDG mirror cache (unexpected file size)
# 2. Yarn GPG key
# 3. Stripe repo (unresolvable host – disable temporarily)
set -e

echo "=== 1. Clearing PGDG apt list cache ==="
sudo rm -f /var/lib/apt/lists/apt.postgresql.org*
echo "Done."

echo ""
echo "=== 2. Adding Yarn APT GPG key ==="
curl -fsSL https://dl.yarnpkg.com/debian/pubkey.gpg | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/yarn.gpg
echo "Done."

echo ""
echo "=== 3. Disabling Stripe repo (packages.stripe.com unresolved) ==="
for f in /etc/apt/sources.list.d/stripe*.list /etc/apt/sources.list.d/*stripe*.list; do
  if [[ -f "$f" ]]; then
    sudo mv "$f" "${f}.disabled" && echo "Disabled $f"
  fi
done
echo "Done. Re-enable later with: sudo mv /etc/apt/sources.list.d/stripe*.list.disabled /etc/apt/sources.list.d/stripe.list"

echo ""
echo "=== 4. Running apt update ==="
sudo apt update

echo ""
echo "=== 5. Installing PostgreSQL 17 and pgvector ==="
sudo apt install -y postgresql-17 postgresql-17-pgvector

echo ""
echo "=== 6. Starting cluster ==="
sudo pg_ctlcluster 17 main start
pg_lsclusters

echo ""
echo "All set. Run: make migrate-local"
