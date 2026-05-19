---
title: Secrets Vault Guide
description: Learn how to use FunctionFly's zero-knowledge secrets vault to securely manage sensitive configuration.
sidebar:
  order: 100
---

The Secrets Vault provides zero-knowledge encryption for sensitive configuration data. Your secrets are encrypted client-side before being stored, ensuring the server never sees plaintext values.

## Quick Start

```bash
# Create a secret
ffly secrets set DB_PASSWORD

# List secrets (values hidden)
ffly secrets list

# Use in function.yaml
ffly secrets set API_KEY="sk-1234567890"
```

## Key Features

- **Zero-knowledge encryption**: Client-side encryption, server never sees plaintext
- **Passphrase protection**: Only you hold the decryption key
- **Runtime injection**: Secrets available at function execution time
- **Scoped access**: Control secrets per function or environment

## Managing Secrets

```bash
# Create with value
ffly secrets set API_KEY="sk-..."

# Create from file
ffly secrets set PRIVATE_KEY --file ./key.pem

# Update
ffly secrets set API_KEY="new-value"

# Delete
ffly secrets delete OLD_KEY
```

## Using in Functions

Reference secrets in `function.yaml`:

```yaml
name: my-function
runtime: python
secrets:
  - DB_PASSWORD
  - API_KEY
environment:
  DATABASE_URL: "postgresql://user:${DB_PASSWORD}@db.example.com/db"
```

Access in code via environment variables - secrets are automatically injected at runtime.

## Setting Your Passphrase

```bash
# Set a new passphrase
ffly secrets passphrase set

# Rotate passphrase (re-encrypts all secrets)
ffly secrets passphrase rotate
```

:::caution
Store your passphrase securely. FunctionFly cannot recover lost passphrases.
:::

## Next Steps

- [Using the Registry](/registry/) - Install functions from the public registry
- [Environment Variables](/guides/environment-variables/) - Configure function settings
- [Rate Limiting](/guides/rate-limiting/) - Manage function usage