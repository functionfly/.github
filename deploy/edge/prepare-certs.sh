#!/bin/bash
# Build fullchain.pem and ensure privkey.pem for edge TLS.
# Put your certificate files in ./certs-in/ (e.g. from Downloads/STAR_functionfly_com):
#   - Leaf (STAR_functionfly_com) and intermediates as .crt/.cer or .pem
#   - Private key as privkey.pem (from your CA or key generation)
# Then run: ./prepare-certs.sh
# Output: ./certs-out/fullchain.pem, ./certs-out/privkey.pem
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERTS_IN="${SCRIPT_DIR}/certs-in"
CERTS_OUT="${SCRIPT_DIR}/certs-out"
CONVERTED="${CERTS_OUT}/.converted"
mkdir -p "$CERTS_IN" "$CERTS_OUT" "$CONVERTED"

# Convert any .cer/.crt (DER or PEM) in certs-in to .pem in .converted
shopt -s nullglob
for f in "${CERTS_IN}"/*.cer "${CERTS_IN}"/*.crt "${CERTS_IN}"/*.pem; do
  [ -f "$f" ] || continue
  base=$(basename "$f" .cer)
  base=$(basename "$base" .crt)
  base=$(basename "$base" .pem)
  out="${CONVERTED}/${base}.pem"
  [ -f "$out" ] && [ "$out" -nt "$f" ] && continue
  if [[ "$f" == *.pem ]]; then
    cp "$f" "$out"
  else
    openssl x509 -in "$f" -out "$out" -inform DER 2>/dev/null || openssl x509 -in "$f" -out "$out" -inform PEM 2>/dev/null || cp "$f" "$out"
  fi
done
shopt -u nullglob 2>/dev/null || true

# If you have a single PEM that already has leaf + intermediates, copy it as fullchain
if [ -f "${CERTS_IN}/fullchain.pem" ]; then
  cp "${CERTS_IN}/fullchain.pem" "${CERTS_OUT}/fullchain.pem"
  echo "Using existing fullchain.pem from certs-in/"
elif [ -f "${CONVERTED}/STAR_functionfly_com.pem" ] || compgen -G "${CONVERTED}/*functionfly*.pem" >/dev/null 2>&1; then
  # Leaf: STAR_functionfly_com or any *functionfly*.pem
  leaf="${CONVERTED}/STAR_functionfly_com.pem"
  [ -f "$leaf" ] || leaf=$(ls "${CONVERTED}"/*functionfly*.pem 2>/dev/null | head -1)
  cat "$leaf" > "${CERTS_OUT}/fullchain.pem"
  # Append intermediates (each other .pem in CONVERTED that looks like a cert)
  for f in "${CONVERTED}"/*.pem; do
    [ -f "$f" ] || continue
    [ "$f" = "$leaf" ] && continue
    grep -q "BEGIN CERTIFICATE" "$f" && cat "$f" >> "${CERTS_OUT}/fullchain.pem"
  done
  echo "Built fullchain.pem from certs-in/"
else
  echo "Put your certs in: ${CERTS_IN}/"
  echo "  - Leaf: STAR_functionfly_com.cer/.crt/.pem (or any *functionfly* cert)"
  echo "  - Intermediates: Sectigo*, SSL2BUY*, USERTrust* (same folder; .cer/.crt/.pem)"
  echo "  - Private key: privkey.pem"
  echo ""
  echo "If filenames are long, copy the folder from Windows (e.g. STAR_functionfly_com) into certs-in/ and run again."
  exit 1
fi

# Private key: optional if keys were generated on the VPS (use KEYS_ON_SERVER=1 when uploading)
if [ -f "${CERTS_IN}/privkey.pem" ]; then
  cp "${CERTS_IN}/privkey.pem" "${CERTS_OUT}/privkey.pem"
  chmod 600 "${CERTS_OUT}/privkey.pem"
  echo "Using private key from certs-in/"
elif [ -f "${CERTS_IN}/STAR_functionfly_com.key" ] || compgen -G "${CERTS_IN}/*.key" >/dev/null 2>&1; then
  key=$(ls "${CERTS_IN}"/*.key 2>/dev/null | head -1)
  cp "$key" "${CERTS_OUT}/privkey.pem"
  chmod 600 "${CERTS_OUT}/privkey.pem"
  echo "Using private key from certs-in/"
else
  echo "No private key in certs-in/ (keys on VPS is OK)."
  echo "Upload only the chain: KEYS_ON_SERVER=1 ./upload-certs.sh"
fi

echo "Ready: ${CERTS_OUT}/fullchain.pem"
[ -f "${CERTS_OUT}/privkey.pem" ] && echo "       ${CERTS_OUT}/privkey.pem"
echo "Next: ./upload-certs.sh  (or KEYS_ON_SERVER=1 ./upload-certs.sh if keys are on each VPS)"
