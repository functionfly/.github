#!/bin/bash
# PostgreSQL SSL Certificate Setup Script
# Run as root on the server before starting PostgreSQL in production
# Usage: ./setup-postgresql-ssl.sh

set -e

SSL_CERT_DIR="/etc/ssl/certs"
SSL_KEY_DIR="/etc/ssl/private"
CERT_FILE="${SSL_CERT_DIR}/postgresql.crt"
KEY_FILE="${SSL_KEY_DIR}/postgresql.key"
CHAIN_FILE="${SSL_CERT_DIR}/postgresql.chain.crt"

echo "=== PostgreSQL SSL Certificate Setup ==="

# Check if certificates already exist
if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
    echo "Certificates already exist at $CERT_FILE and $KEY_FILE"
    echo "To regenerate, remove them first:"
    echo "  rm $CERT_FILE $KEY_FILE"
    exit 0
fi

# Create directories if they don't exist
mkdir -p "$SSL_CERT_DIR" "$SSL_KEY_DIR"
chmod 710 "$SSL_KEY_DIR"

# Generate self-signed certificate for immediate use
# In production, replace with certificates from Let's Encrypt or your CA
echo "Generating self-signed certificate for PostgreSQL..."
openssl req -new -x509 -days 365 -nodes \
    -text \
    -out "$CERT_FILE" \
    -keyout "$KEY_FILE" \
    -subj "/C=US/ST=State/L=City/O=FunctionFly/CN=postgres"

# Set permissions
chmod 644 "$CERT_FILE"
chmod 600 "$KEY_FILE"

# If Let's Encrypt certificate exists, create certificate chain
if [ -f "/etc/letsencrypt/live/$(hostname)/fullchain.pem" ]; then
    echo "Creating certificate chain from Let's Encrypt..."
    cp "/etc/letsencrypt/live/$(hostname)/fullchain.pem" "$CHAIN_FILE"
    cp "/etc/letsencrypt/live/$(hostname)/privkey.pem" "$KEY_FILE"
    chmod 644 "$CHAIN_FILE"
    chmod 600 "$KEY_FILE"
fi

echo "=== SSL Certificate Setup Complete ==="
echo "Certificate: $CERT_FILE"
echo "Key: $KEY_FILE"
echo ""
echo "Remember to restart PostgreSQL after certificate setup:"
echo "  sudo systemctl restart postgresql"
