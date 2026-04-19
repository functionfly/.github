---
title: Rate Limiting
description: Configure rate limits and usage quotas for your functions.
sidebar:
  order: 5
---

Control access to your functions with configurable rate limits and usage quotas.

## Overview

Rate limiting protects your functions from:

- **Abuse**: Prevent malicious or accidental overuse
- **Cost control**: Manage your execution costs
- **Fair usage**: Ensure consistent performance for all users
- **Protection**: Guard against DDoS attacks

## Rate Limit Types

### 1. Per-Function Limits

Set limits on individual functions:

```yaml
# function.yaml
name: my-function
runtime: python
rate_limits:
  requests_per_minute: 60
  burst: 10
  daily_quota: 10000
```

### 2. Per-User Limits

Control access per API key or user:

```yaml
rate_limits:
  per_user:
    requests_per_minute: 10
    daily_quota: 1000
```

### 3. Global Limits

Account-wide rate limiting:

```yaml
rate_limits:
  global:
    requests_per_minute: 1000
    concurrent: 100
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `requests_per_minute` | Requests allowed per minute | 60 |
| `burst` | Burst capacity for short spikes | 10 |
| `daily_quota` | Maximum requests per day | Unlimited |
| `concurrent` | Simultaneous executions | 100 |
| `cooldown` | Seconds between requests from same source | 0 |

## Rate Limit Headers

Every response includes rate limit information:

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1640995200
X-RateLimit-Retry-After: 30
```

### Header Descriptions

- **X-RateLimit-Limit**: Maximum requests allowed per window
- **X-RateLimit-Remaining**: Requests remaining in current window
- **X-RateLimit-Reset**: Unix timestamp when limit resets
- **X-RateLimit-Retry-After**: Seconds until you can retry (only on 429 responses)

## Handling Rate Limit Errors

When rate limits are exceeded, the API returns:

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json

{
  "error": "rate_limit_exceeded",
  "message": "Rate limit exceeded. Try again in 30 seconds.",
  "retry_after": 30
}
```

### Retry Strategies

**Exponential Backoff:**

```python
import time
import random

def invoke_with_retry(func, max_retries=3):
    for attempt in range(max_retries):
        try:
            return func()
        except RateLimitError as e:
            if attempt == max_retries - 1:
                raise
            
            # Exponential backoff with jitter
            delay = (2 ** attempt) + random.random()
            time.sleep(delay)
```

## Usage Quotas

### Daily Limits

```yaml
# function.yaml
quotas:
  daily:
    requests: 10000
    compute_seconds: 3600  # 1 hour of execution time
    egress_bytes: 104857600  # 100 MB outbound data
```

### Monthly Limits

```yaml
quotas:
  monthly:
    requests: 100000
    compute_seconds: 36000
```

### Custom Time Windows

```yaml
quotas:
  custom:
    window: 1h  # 1 hour
    requests: 500
```

## Alerting

Configure alerts for quota thresholds:

```bash
# Set up quota alerts
fly limits alerts set \
  --function my-function \
  --daily-requests 80% \
  --email admin@example.com
```

### Alert Channels

- **Email**: Sent to account owner
- **Webhook**: POST to your endpoint
- **Dashboard**: In-app notifications
- **Slack**: Direct integration

## Plan Limits

Different plans include different rate limits:

| Plan | Requests/Min | Burst | Daily Quota |
|------|-------------|-------|-------------|
| Free | 60 | 10 | 10,000 |
| Pro | 600 | 100 | 100,000 |
| Business | 3000 | 500 | 500,000 |
| Enterprise | Custom | Custom | Unlimited |

## Advanced Configuration

### Path-Specific Limits

Different limits for different endpoints:

```yaml
rate_limits:
  paths:
    "/webhook":
      requests_per_minute: 120
      burst: 20
    "/api/v1/data":
      requests_per_minute: 30
      daily_quota: 5000
```

### Method-Specific Limits

Different limits for HTTP methods:

```yaml
rate_limits:
  methods:
    GET:
      requests_per_minute: 100
    POST:
      requests_per_minute: 30
    DELETE:
      requests_per_minute: 10
```

### IP-Based Limits

Restrict by IP address:

```yaml
rate_limits:
  ip_whitelist:
    - 192.168.1.0/24
    - 10.0.0.50
  ip_blacklist:
    - 192.168.1.100
```

## Monitoring

View rate limit usage in the dashboard:

```bash
# Get current usage stats
fly limits stats

# Get usage for a specific function
fly limits stats --function my-function

# Historical data
fly limits stats --from 2024-01-01 --to 2024-01-31
```

## Best Practices

1. **Set conservative limits**: Start low and increase as needed
2. **Monitor usage**: Regularly review your rate limit dashboards
3. **Handle 429s**: Implement retry logic in your clients
4. **Separate by environment**: Different limits for dev/staging/prod
5. **Document limits**: Tell your users about rate limits

## Troubleshooting

### Unexpected 429s

- Check if you're sharing an API key across multiple clients
- Verify the correct rate limits are configured
- Look for retry storms (cascading retries)

### Limits not applying

- Ensure the function.yaml is deployed
- Verify rate_limits section syntax
- Check for conflicting global/function limits

## Next Steps

- Learn about [webhooks](/guides/webhooks/) for event-driven integrations
- Set up [CI/CD pipelines](/guides/ci-cd/) for automated deployments
- Explore [monitoring and analytics](/analytics/) for insights
