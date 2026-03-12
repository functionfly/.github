#!/bin/bash
# Generate a new private key and CSR for reissuing the wildcard cert.
# You will submit the CSR to your CA (e.g. SSL2BUY/Sectigo); they send back
# a new zip with four certs. Then use the same key + new certs to deploy.
#
# Usage: ./reissue-csr.sh
# Output: certs-out/privkey.pem (new key) and certs-out/functionfly.csr (submit to CA)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERTS_OUT="${SCRIPT_DIR}/certs-out"
mkdir -p "$CERTS_OUT"

KEY="${CERTS_OUT}/privkey.pem"
CSR="${CERTS_OUT}/functionfly.csr"

if [ -f "$KEY" ]; then
  echo "Found existing privkey.pem in certs-out."
  read -p "Overwrite and generate a new key? (y/N): " confirm
  if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "Aborted. Using existing key for CSR."
  else
    rm -f "$KEY" "$CSR"
  fi
fi

if [ ! -f "$KEY" ]; then
  echo "Generating new 2048-bit RSA key..."
  openssl genrsa -out "$KEY" 2048
  chmod 600 "$KEY"
fi

echo "Generating CSR for *.functionfly.com (CN only; no SAN - some CAs reject CSR with SAN)..."
openssl req -new -key "$KEY" -out "$CSR" -subj "/CN=*.functionfly.com"

echo ""
echo "Done."
echo "  Private key: $KEY  (keep this; you will deploy it to both VPS with the new cert)"
echo "  CSR:          $CSR"
echo ""
echo "Next steps:"
echo "  1. Open $CSR and copy its contents (including -----BEGIN/END CERTIFICATE REQUEST-----)."
echo "  2. Log in to your CA (e.g. SSL2BUY/Sectigo), request reissue for *.functionfly.com, paste the CSR."
echo "  3. Download the new certificate zip (four certs)."
echo "  4. Put the four certs in certs-in/ (extract zip there). privkey.pem is already in certs-out/."
echo "  5. Run: ./prepare-certs.sh   (builds fullchain.pem; will use certs-out/privkey.pem if certs-in has no key)"
echo "  6. Run: ./upload-certs.sh   (uploads fullchain.pem AND privkey.pem to both VPS — do NOT use KEYS_ON_SERVER)"
echo "  7. On each VPS: chown root:caddy /etc/ssl/functionfly/privkey.pem /etc/ssl/functionfly/fullchain.pem && chmod 640 /etc/ssl/functionfly/privkey.pem && systemctl restart caddy"
echo ""
