#!/bin/bash
# One-shot: prepare certs from certs-in/ and import the wildcard into Fly.io for api.functionfly.com.
#
# 1. Extract your CA zip (e.g. 2891200744.zip) into deploy/edge/certs-in/
# 2. Add your private key to certs-in/ as privkey.pem (or .key) if the CA didn't include it
# 3. From repo root:  deploy/edge/import-fly-cert.sh
#    Or from deploy/edge:  ./import-fly-cert.sh
#
# Optional: FLY_APP=myapp FLY_HOSTNAME=api.mydomain.com ./import-fly-cert.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

FLY_APP="${FLY_APP:-functionfly-control}"
FLY_HOSTNAME="${FLY_HOSTNAME:-api.functionfly.com}"

CERTS_OUT="${SCRIPT_DIR}/certs-out"
if [ -f "${CERTS_OUT}/fullchain.pem" ] && [ -f "${CERTS_OUT}/privkey.pem" ]; then
  echo "Using existing fullchain.pem and privkey.pem from certs-out/"
else
  echo "Preparing fullchain.pem and privkey.pem from certs-in/..."
  ./prepare-certs.sh
fi

if [ ! -f "${CERTS_OUT}/fullchain.pem" ] || [ ! -f "${CERTS_OUT}/privkey.pem" ]; then
  echo "Need both fullchain.pem and privkey.pem in certs-out/ for Fly import."
  echo "Add your private key to certs-in/ as privkey.pem or .key, then run again."
  exit 1
fi

echo "Importing certificate to Fly.io (app=$FLY_APP, hostname=$FLY_HOSTNAME)..."
flyctl certs import "$FLY_HOSTNAME" --app "$FLY_APP" \
  --fullchain "${CERTS_OUT}/fullchain.pem" \
  --private-key "${CERTS_OUT}/privkey.pem"

echo "Done. Verify with: flyctl certs check $FLY_HOSTNAME --app $FLY_APP && curl -sI https://$FLY_HOSTNAME/healthz"
