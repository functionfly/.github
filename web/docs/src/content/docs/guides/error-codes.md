---
title: Error Codes & Troubleshooting
description: Common errors, their meanings, and how to resolve them.
sidebar:
  order: 11
---



This guide covers common errors you may encounter with FunctionFly and how to resolve them.

## HTTP Status Codes

### 2xx Success

| Code | Meaning | Resolution |
|------|---------|------------|
| `200 OK` | Request succeeded | No action needed |
| `201 Created` | Resource created successfully | Resource is ready to use |
| `204 No Content` | Request succeeded with no response body | Check for expected empty response |

### 4xx Client Errors

| Code | Error | Common Causes | Resolution |
|------|-------|---------------|------------|
| `400 Bad Request` | Invalid request format | Malformed JSON, missing required fields | Check request body structure |
| `401 Unauthorized` | Authentication required | Missing or invalid API key | Verify `FFLY_TOKEN` or API key |
| `403 Forbidden` | Insufficient permissions | Valid auth but lacking permissions | Check account tier limits |
| `404 Not Found` | Resource doesn't exist | Typo in function name, deleted resource | Verify resource exists |
| `409 Conflict` | Resource conflict | Function already exists, version conflict | Use different name or version |
| `422 Unprocessable Entity` | Validation failed | Invalid field values | Check error details in response |
| `429 Too Many Requests` | Rate limit exceeded | Too many requests per minute | Implement exponential backoff |

### 5xx Server Errors

| Code | Error | Common Causes | Resolution |
|------|-------|---------------|------------|
| `500 Internal Server Error` | Server-side error | Unexpected condition | Check status page, retry later |
| `502 Bad Gateway` | Upstream error | Function runtime failure | Check function logs |
| `503 Service Unavailable` | Temporary outage | Maintenance, overload | Check status page |
| `504 Gateway Timeout` | Upstream timeout | Slow function execution | Optimize function or increase timeout |

---

## Function Execution Errors

### `ERR_FUNCTION_NOT_FOUND`

**Message:** `Function "{name}" not found in registry`

**Cause:** Function doesn't exist or hasn't been published

**Solutions:**
```bash
# Verify function exists
ffly list

# Publish the function
ffly publish

# Or deploy a new version
ffly deploy --version 1.0.1
```

### `ERR_RUNTIME_NOT_SUPPORTED`

**Message:** `Runtime "{runtime}" is not supported`

**Cause:** Using an unsupported language runtime

**Solutions:**
- Supported runtimes: `python`, `nodejs`, `javascript`, `go`, `rust`, `wasm`, `ruby`, `kotlin`, `swift`
- Check `functionfly.jsonc` for typos
- Use `--runtime` flag to specify correct runtime

### `ERR_TIMEOUT_EXCEEDED`

**Message:** `Function execution timeout after {ms}ms`

**Cause:** Function took longer than the configured timeout

**Solutions:**
```bash
# Increase timeout in functionfly.jsonc
{
    "timeout_ms": 30000  // 30 seconds
}

# Or via CLI
ffly deploy --timeout 60000
```

### `ERR_MEMORY_LIMIT_EXCEEDED`

**Message:** `Memory limit of {limit}MB exceeded`

**Cause:** Function used more memory than allowed

**Solutions:**
- Optimize memory usage in your function
- Increase memory limit (if on Professional+):
```bash
ffly deploy --memory 512  # MB
```

### `ERR_INVALID_RESPONSE`

**Message:** `Function returned invalid response format`

**Cause:** Function didn't return a valid response structure

**Expected format:**
```json
{
    "statusCode": 200,
    "body": "response data"
}
```

---

## Authentication Errors

### `ERR_INVALID_API_KEY`

**Message:** `Invalid API key`

**Cause:** API key is malformed, expired, or revoked

**Solutions:**
```bash
# Regenerate API key in dashboard
# Settings → API Keys → Create New Key

# Or login again
ffly logout
ffly login
```

### `ERR_TOKEN_EXPIRED`

**Message:** `Session token has expired`

**Cause:** Refresh token has expired (typically after 30 days of inactivity)

**Solutions:**
```bash
# Re-authenticate
ffly logout
ffly login
```

### `ERR_HMAC_SIGNATURE_INVALID`

**Message:** `HMAC signature verification failed`

**Cause:** Request was modified after signing, or shared secret mismatch

**Solutions:**
- Verify `FUNCTIONFLY_API_SECRET` matches dashboard
- Ensure request body isn't modified after signing
- Check clock synchronization on client/server

---

## Deployment Errors

### `ERR_BUILD_FAILED`

**Message:** `Build failed: {details}`

**Common causes and solutions:**

| Cause | Solution |
|-------|----------|
| Syntax errors in code | Fix code and redeploy |
| Missing dependencies | Add `requirements.txt` or `package.json` |
| Invalid function signature | Ensure export default handler exists |
| Platform-specific binary | Use multi-platform Docker base image |

```bash
# Debug locally first
ffly dev

# Verbose output
ffly deploy --verbose
```

### `ERR_QUOTA_EXCEEDED`

**Message:** `{resource} quota exceeded for {plan} plan`

**Cause:** You've hit a plan limit

**Solutions:**
- Free tier: 3 functions, 25K requests/month
- Upgrade to Starter/Professional/Enterprise
- Delete unused functions to free quota

### `ERR_CANNOT_DELETE_FUNCTION`

**Message:** `Function has active subscriptions`

**Cause:** Other users are still using this function

**Solutions:**
- Wait for subscriptions to expire (30 days)
- Contact support to force-delete

---

## StateFabric Errors

### `ERR_STATE_OBJECT_NOT_FOUND`

**Message:** `State object "{name}" not found`

**Cause:** State object doesn't exist

**Solutions:**
```javascript
// Create state object first
const state = await context.state.create('my-cart', {
    initialData: { items: [] }
});
```

### `ERR_STATE_TRANSACTION_CONFLICT`

**Message:** `Concurrent modification detected`

**Cause:** Two executions tried to modify state simultaneously

**Solutions:**
- Use optimistic locking
- Retry the transaction
```javascript
const state = await context.state.get('my-cart', { 
    ifVersion: currentVersion 
});
// If conflict, retry with fresh data
```

### `ERR_STATE_SNAPSHOT_CORRUPT`

**Message:** `State snapshot verification failed`

**Cause:** Data corruption in snapshot

**Solutions:**
- Restore from event replay
- Contact support for manual recovery
- Use `statefabric.repair()` API

---

## Agent Errors

### `ERR_AGENT_NOT_INITIALIZED`

**Message:** `Agent not initialized or initialization failed`

**Cause:** Agent dependencies not properly set up

**Solutions:**
```bash
# Verify agent configuration
ffly agents verify

# Re-initialize agent
ffly agents init my-agent --force
```

### `ERR_AGENT_LOOP_DETECTED`

**Message:** `Infinite loop detected in agent reasoning`

**Cause:** Agent is stuck in repetitive thought patterns

**Solutions:**
- Simplify agent instructions
- Add explicit termination conditions
- Reduce `max_iterations` in policy

### `ERR_TRUST_VERIFICATION_FAILED`

**Message:** `Trust verification failed: score {score} below threshold`

**Cause:** Function/agent trust score is too low

**Solutions:**
- Review and improve function quality
- Complete trust verification process
- Use `trust.verify()` API to check current score

---

## Network & Connection Errors

### `ERR_CONNECTION_REFUSED`

**Message:** `Connection refused to {host}:{port}`

**Cause:** Service is down or firewall blocking

**Solutions:**
```bash
# Check service status
curl https://api.functionfly.com/health

# Verify network connectivity
ping api.functionfly.com

# Check firewall rules
sudo ufw status
```

### `ERR_SSL_CERTIFICATE_INVALID`

**Message:** `SSL certificate verification failed`

**Cause:** Expired, self-signed, or untrusted certificate

**Solutions:**
- Update system CA certificates: `sudo apt update && sudo apt upgrade ca-certificates`
- For local development, disable SSL verification:
```bash
export NODE_TLS_REJECT_UNAUTHORIZED=0  # DEV ONLY
```

### `ERR_DNS_RESOLUTION_FAILED`

**Message:** `Could not resolve host: {hostname}`

**Cause:** DNS configuration issue

**Solutions:**
```bash
# Flush DNS cache
sudo systemd-resolve --flush-caches

# Or use Google DNS
echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf
```

---

## Data & Storage Errors

### `ERR_DATABASE_CONNECTION_FAILED`

**Message:** `Could not connect to database`

**Cause:** Wrong credentials, database down, or network issue

**Solutions:**
```bash
# Test connection
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME

# Check DATABASE_URL format
# Should be: postgres://user:pass@host:5432/dbname?sslmode=require
```

### `ERR_STORAGE_QUOTA_EXCEEDED`

**Message:** `Storage quota of {limit}GB exceeded`

**Cause:** Too much data stored

**Solutions:**
- Clean up old function versions: `ffly cleanup --versions`
- Delete unused assets: `ffly storage ls`
- Increase storage limit by upgrading plan

---

## Rate Limiting

### Understanding Rate Limits

| Plan | Requests/Minute | Burst |
|------|-----------------|-------|
| Free | 60 | 10 |
| Starter | 300 | 50 |
| Professional | 1000 | 200 |
| Enterprise | 5000+ | Custom |

### Implementing Backoff

```javascript
// JavaScript example
async function fetchWithRetry(url, retries = 3) {
    for (let i = 0; i < retries; i++) {
        try {
            const response = await fetch(url);
            if (response.status === 429) {
                const wait = Math.pow(2, i) * 1000;
                await new Promise(r => setTimeout(r, wait));
                continue;
            }
            return response;
        } catch (err) {
            if (i === retries - 1) throw err;
        }
    }
}
```

---

## Getting Help

If you're stuck:

1. **Check the [status page](https://status.functionfly.com)** for outages
2. **Search [GitHub Discussions](https://github.com/functionfly/functionfly/discussions)** for similar issues
3. **Ask in [Discord](https://discord.gg/FGnseK9Bp)** — invite-only launch this month
4. **Contact support** with your `REQUEST_ID` from error responses

When contacting support, include:
- `REQUEST_ID` from the error response
- Timestamp of the error
- Function name and version
- Full error message
- Steps to reproduce
