---
title: Secrets Vault Guide
description: Manage sensitive configuration with FunctionFly's zero-knowledge secrets vault
sidebar:
  order: 100
---

The Secrets Vault provides zero-knowledge encryption for sensitive configuration data. Your secrets are encrypted client-side before being stored, ensuring the server never sees plaintext values.

## Quick Start

```bash
# Create a secret (prompts for value)
ff secrets set DB_PASSWORD

# Create with inline value
ff secrets set API_KEY="sk-1234567890"

# List secrets (values hidden)
ff secrets list

# Delete a secret
ff secrets unset OLD_KEY
```

## Key Features

- **Zero-knowledge encryption** — Client-side encryption, server never sees plaintext
- **Passphrase protection** — Only you hold the decryption key
- **Runtime injection** — Secrets available at function execution time
- **Scoped access** — Control secrets per function or environment

## Managing Secrets

```bash
# Set a secret (prompts for value if not provided inline)
ff secrets set API_KEY="sk-..."

# Set from file
ff secrets set PRIVATE_KEY --file ./key.pem

# Update an existing secret
ff secrets set API_KEY="new-value"

# Delete a secret
ff secrets unset OLD_KEY

# List all secrets (names only, values never shown)
ff secrets list
```

## Environment Variables

The CLI also manages non-secret environment variables:

```bash
# List env vars
ff env list

# Set an env var
ff env set NODE_ENV=production

# Get a specific var
ff env get NODE_ENV

# Unset an env var
ff env unset NODE_ENV
```

## Using Secrets in Functions

Reference secrets in `functionfly.jsonc`:

```jsonc
{
  "name": "my-function",
  "version": "1.0.0",
  "runtime": "node20",
  "env": {
    "DATABASE_URL": "postgresql://user:${DB_PASSWORD}@db.example.com/db"
  }
}
```

Secrets are automatically injected as environment variables at runtime. Access them in code:

```javascript
// Node.js
const apiKey = process.env.API_KEY;
```

```python
# Python
import os
api_key = os.environ["API_KEY"]
```

## Setting Your Passphrase

The passphrase is the master key for client-side encryption. It is never
sent to FunctionFly's servers.

```bash
# Set a new passphrase (interactive prompt)
ff secrets passphrase set

# Rotate passphrase (re-encrypts all secrets)
ff secrets passphrase rotate
```

:::caution
Store your passphrase securely. FunctionFly cannot recover lost passphrases.
:::

## Vault Plans

Vault features are included with your platform plan:

| Platform Plan | Vault Tier | Max Secrets | Max Dynamic Creds | Tokens/Secret | Key Features |
|---------------|-----------|------------|-------------------|---------------|--------------|
| **Free** | Free | 25 | 100 | 5 | Expiration, quota widget |
| **Starter** ($24/mo) | Free | 25 | 100 | 5 | Same as Free |
| **Professional** ($79/mo) | Pro | 500 | 5,000 | 25 | MFA, IP allowlist, breakGlass, namespaces, auditExport, tokenMonitor, rotationSchedules |
| **Enterprise** ($299/mo) | Team | 5,000 | 50,000 | 100 | Everything in Pro + escrow, RBAC, shares, siemWebhooks, dependencyGraph |
| **Agent Enterprise** ($499/mo) | Enterprise | 1,000,000 | 1,000,000 | 1,000 | Everything in Team + SSO, HA status |

See [Secrets Vault](/secrets-vault/) for full vault documentation.

## Next Steps

- [Environment Variables](/guides/environment-variables/) — Configure function settings
- [Secrets Vault](/secrets-vault/) — Vault overview and architecture
- [CLI Reference](/cli/) — Full CLI documentation
- [API Keys](/api-keys/) — Managing API keys
