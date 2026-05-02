---
title: Secrets & Vault
description: Store and manage sensitive configuration with zero-knowledge encryption.
sidebar:
  order: 4
---

FunctionFly's Secrets Vault provides zero-knowledge encryption for sensitive configuration data. Your secrets are encrypted client-side before being stored, ensuring the server never sees plaintext values.

## How It Works

1. **Client-Side Encryption**: Secrets are encrypted in your browser or CLI before being sent
2. **Passphrase Protection**: Only you hold the decryption passphrase
3. **Secure Storage**: Encrypted ciphertext is stored with no server-side access
4. **Runtime Injection**: Decrypted secrets are injected at execution time

## Creating Secrets

### Via Dashboard

1. Navigate to **Secrets & Vault** in the dashboard
2. Click **New Secret**
3. Enter a name and value
4. Set your encryption passphrase (or use a generated one)
5. Save the encrypted secret

### Via CLI

```bash
# Create a secret interactively
ffly secrets set DB_PASSWORD

# Create a secret with a value
ffly secrets set API_KEY="sk-1234567890"

# Create from a file
ffly secrets set PRIVATE_KEY --file ./key.pem

# Create from environment variable
ffly secrets set WEBHOOK_SECRET --env
```

## Using Secrets in Functions

### Reference in function.yaml

```yaml
name: my-function
runtime: python
secrets:
  - DB_PASSWORD
  - API_KEY
environment:
  DATABASE_URL: "postgresql://user:${DB_PASSWORD}@db.example.com/db"
```

### Access in Code

**Python:**
```python
import os

def handler(request):
    api_key = os.environ.get('API_KEY')
    # Use the secret...
```

**JavaScript:**
```javascript
export default async function handler(request) {
  const apiKey = process.env.API_KEY;
  // Use the secret...
}
```

**Go:**
```go
func Handler(w http.ResponseWriter, r *http.Request) {
    apiKey := os.Getenv("API_KEY")
    // Use the secret...
}
```

## Encryption Passphrase

### Setting Your Passphrase

The passphrase is used to encrypt and decrypt your secrets:

```bash
# Set a new passphrase
ffly secrets passphrase set

# Change existing passphrase (re-encrypts all secrets)
ffly secrets passphrase rotate
```

### Recovery

:::caution
FunctionFly cannot recover lost passphrases. Store your passphrase securely:

- Password manager (1Password, Bitwarden, etc.)
- Hardware security key
- Secure offline storage
:::

### Passphrase Requirements

- Minimum 12 characters
- Must include uppercase, lowercase, and numbers
- Special characters recommended
- Avoid dictionary words

## Managing Secrets

### List Secrets

```bash
# List all secret names (values are never shown)
ffly secrets list

# List with metadata
ffly secrets list --verbose
```

### Update Secrets

```bash
# Update an existing secret
ffly secrets set API_KEY="new-value"

# Update multiple secrets at once
ffly secrets set DB_USER="admin" DB_PASS="secret"
```

### Delete Secrets

```bash
# Remove a secret
ffly secrets delete API_KEY

# Bulk delete
ffly secrets delete OLD_KEY1 OLD_KEY2
```

## Secret Scopes

### Function-Specific Secrets

```bash
# Secret only available to a specific function
ffly secrets set DATABASE_URL --function my-function
```

### Environment-Specific Secrets

```bash
# Secret only for production environment
ffly secrets set STRIPE_KEY --environment production
```

### Global Secrets

```bash
# Secret available to all functions (default)
ffly secrets set SHARED_CONFIG
```

## Best Practices

1. **Separate secrets by environment**: Don't reuse production secrets in development
2. **Rotate regularly**: Change secrets every 90 days
3. **Use least privilege**: Only expose needed secrets to each function
4. **Audit access**: Review which functions use which secrets
5. **Avoid logging**: Never log secret values

## Security Model

### Zero-Knowledge Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   Server    │     │   Runtime   │
│  (Encrypt)  │     │ (Encrypted) │────▶│  (Inject)   │
└─────────────┘     └─────────────┘     └─────────────┘
      │                                     │
      │         Passphrase (Client-Only)    │
      └─────────────────────────────────────┘
```

- Server stores only encrypted ciphertext
- Passphrase never leaves the client
- Runtime receives injected values, not the passphrase

### Audit Logging

All secret operations are logged for security:

- Secret creation (name only)
- Secret updates (name only)
- Secret deletions
- Access by functions
- Passphrase rotations

:::note
Secret values are never logged or accessible to FunctionFly staff.
:::

## Troubleshooting

### "Unable to decrypt secret"

- Verify you're using the correct passphrase
- Check that the secret was created with the current passphrase
- If passphrase was rotated, secrets need to be re-encrypted

### "Secret not found"

- Verify the secret name (case-sensitive)
- Check that the secret is in the correct scope (function/environment)
- Ensure the secret hasn't been deleted

### Runtime errors

- Confirm the secret is referenced in function.yaml
- Check that the function has permission to access the secret
- Verify the secret name matches between definition and code

## Next Steps

- Learn about [rate limiting](/guides/rate-limiting/) to manage function usage
- Set up [webhooks](/guides/webhooks/) for event-driven integrations
- Explore [CI/CD integration](/guides/ci-cd/) for automated deployments
