---
title: Authentication & API Keys
description: Set up authentication for your functions and manage API keys securely.
sidebar:
  order: 3
---

FunctionFly supports multiple authentication methods to secure your functions and API access.

## Authentication Methods

### 1. Session-Based Authentication

For browser-based applications and the dashboard:

```bash
# Login via CLI
ffly login

# This opens a browser window for authentication
# and stores a session token locally
```

### 2. API Keys

For programmatic access and server-to-server communication:

```bash
# Create a new API key
ffly keys create --name "Production API" \
  --permissions read,execute \
  --environments production

# List your API keys
ffly keys list

# Revoke an API key
ffly keys revoke <key-id>
```

### 3. JWT Tokens

For temporary access or delegated authentication:

```bash
# Generate a JWT token
ffly token generate --expires 24h

# Use the token in API requests
curl https://api.functionfly.com/v1/execute/<function-id> \
  -H "Authorization: Bearer <jwt-token>"
```

## Securing Functions

### Public Functions

By default, functions are publicly accessible:

```yaml
# function.yaml
name: my-function
runtime: python
visibility: public  # Anyone can invoke
```

### Authenticated Functions

Require authentication to invoke:

```yaml
# function.yaml
name: my-function
runtime: python
visibility: authenticated  # Requires valid API key or session
```

### Private Functions

Only accessible by the owner:

```yaml
# function.yaml
name: my-function
runtime: python
visibility: private  # Only you can invoke
```

## Using API Keys

### In Headers

```bash
curl https://api.functionfly.com/v1/execute/<function-id> \
  -H "Authorization: Bearer <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{"input": "data"}'
```

### In SDKs

**Python:**
```python
from functionfly import Client

client = Client(api_key="your-api-key")
result = client.invoke("function-name", {"input": "data"})
```

**JavaScript:**
```javascript
import { Client } from '@functionfly/sdk';

const client = new Client({ apiKey: 'your-api-key' });
const result = await client.invoke('function-name', { input: 'data' });
```

**Go:**
```go
import "github.com/functionfly/functionfly/sdk/go"

client := functionfly.NewClient(functionfly.WithAPIKey("your-api-key"))
result, err := client.Invoke(ctx, "function-name", map[string]interface{}{
    "input": "data",
})
```

## API Key Permissions

Control what each key can do:

| Permission | Description |
|------------|-------------|
| `read` | View functions and execution history |
| `execute` | Invoke functions |
| `deploy` | Deploy and update functions |
| `admin` | Full account access |

### Example: Read-Only Key

```bash
ffly keys create \
  --name "Monitoring Service" \
  --permissions read \
  --environments production
```

### Example: Deployment Key

```bash
ffly keys create \
  --name "CI/CD Pipeline" \
  --permissions read,deploy \
  --environments staging,production
```

## Environment Scoping

Restrict API keys to specific environments:

```bash
# Development only
ffly keys create --name "Dev Key" --environments development

# Staging and production
ffly keys create --name "Prod Key" --environments staging,production
```

## Rate Limits

API keys have built-in rate limits:

| Plan | Requests/Min | Burst |
|------|-------------|-------|
| Free | 60 | 10 |
| Pro | 600 | 100 |
| Enterprise | Custom | Custom |

## Rotating Keys

Regularly rotate API keys for security:

```bash
# Create a new key with the same permissions
ffly keys rotate <old-key-id> --name "Production API (Rotated)"

# The old key remains active for 24 hours
# Update your applications with the new key
# Then revoke the old key
ffly keys revoke <old-key-id>
```

## Best Practices

1. **Use environment-specific keys**: Separate keys for dev, staging, and production
2. **Principle of least privilege**: Grant only necessary permissions
3. **Rotate regularly**: Change keys every 90 days
4. **Store securely**: Use environment variables or secret managers
5. **Monitor usage**: Review key usage in the dashboard

## Troubleshooting

### 401 Unauthorized

- Check that the API key is valid and not expired
- Verify the key has the required permissions
- Ensure the Authorization header format is correct

### 403 Forbidden

- The key may not have permission for the requested action
- The function may be private or require different authentication

## Next Steps

- Learn about [Secrets & Vault](/guides/secrets-vault/) for managing sensitive data
- Set up [rate limiting](/guides/rate-limiting/) to protect your functions
- Explore [webhooks](/guides/webhooks/) for event-driven workflows
