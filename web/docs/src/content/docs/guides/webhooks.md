---
title: Webhooks
description: Set up and handle webhook events from your functions.
sidebar:
  order: 6
---

Webhooks enable event-driven workflows by allowing your functions to receive HTTP callbacks when events occur.

## What Are Webhooks?

Webhooks are HTTP callbacks triggered by events:

- **Outgoing**: Your function sends webhooks to external services
- **Incoming**: External services send webhooks to your function
- **Internal**: Function-to-function communication via events

## Receiving Webhooks

### Creating a Webhook Endpoint

Create a function that accepts webhook requests:

**Python:**
```python
# webhook_handler.py
import json
import hmac
import hashlib

def handler(request):
    # Verify webhook signature
    signature = request.headers.get('X-Webhook-Signature')
    payload = request.body
    
    if not verify_signature(payload, signature):
        return {
            'statusCode': 401,
            'body': json.dumps({'error': 'Invalid signature'})
        }
    
    # Process the webhook
    event = json.loads(payload)
    process_event(event)
    
    return {
        'statusCode': 200,
        'body': json.dumps({'received': True})
    }

def verify_signature(payload, signature):
    expected = hmac.new(
        WEBHOOK_SECRET.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f'sha256={expected}', signature)
```

**JavaScript:**
```javascript
// webhook-handler.js
import crypto from 'crypto';

export default async function handler(request) {
  // Verify webhook signature
  const signature = request.headers['x-webhook-signature'];
  
  if (!verifySignature(request.body, signature)) {
    return {
      statusCode: 401,
      body: JSON.stringify({ error: 'Invalid signature' })
    };
  }
  
  // Process the webhook
  const event = JSON.parse(request.body);
  await processEvent(event);
  
  return {
    statusCode: 200,
    body: JSON.stringify({ received: true })
  };
}

function verifySignature(payload, signature) {
  const expected = crypto
    .createHmac('sha256', process.env.WEBHOOK_SECRET)
    .update(payload)
    .digest('hex');
  return crypto.timingSafeEqual(
    Buffer.from(`sha256=${expected}`),
    Buffer.from(signature)
  );
}
```

### Webhook Configuration

```yaml
# function.yaml
name: webhook-handler
runtime: python
webhook:
  path: /webhooks/events
  methods: [POST]
  verify_signature: true
  secret: ${WEBHOOK_SECRET}
secrets:
  - WEBHOOK_SECRET
```

### Deploying Webhook Functions

```bash
# Deploy the webhook handler
ffly deploy

# Get the webhook URL
ffly info

# URL will be: https://api.functionfly.com/v1/execute/<function-id>/webhooks/events
```

## Sending Webhooks

### Outgoing Webhooks

Send webhooks from your function to external services:

```python
# notify.py
import requests
import json

def handler(request):
    # Process some data
    result = process_data(request)
    
    # Send webhook to external service
    webhook_url = "https://example.com/webhooks/functionfly"
    
    response = requests.post(
        webhook_url,
        json={
            'event': 'processing_complete',
            'data': result,
            'timestamp': datetime.utcnow().isoformat()
        },
        headers={
            'Content-Type': 'application/json',
            'X-FunctionFly-Signature': generate_signature(result)
        },
        timeout=30
    )
    
    return {
        'statusCode': 200,
        'body': json.dumps({
            'webhook_sent': response.status_code == 200
        })
    }
```

### Webhook Retries

FunctionFly automatically retries failed webhooks:

```yaml
# function.yaml
webhook:
  retry_policy:
    max_attempts: 5
    backoff: exponential
    initial_delay: 1s
    max_delay: 60s
```

## Event Types

### Standard Events

| Event | Description |
|-------|-------------|
| `function.invoked` | Function was invoked |
| `function.completed` | Function execution completed |
| `function.failed` | Function execution failed |
| `deployment.success` | Function deployed successfully |
| `deployment.failed` | Function deployment failed |

### Custom Events

Define your own event types:

```yaml
# function.yaml
events:
  types:
    - user.created
    - order.processed
    - payment.received
```

### Subscribing to Events

```bash
# Subscribe to an event
ffly events subscribe function.completed \
  --webhook https://my-function.com/webhooks/completed

# Subscribe with filtering
ffly events subscribe user.created \
  --webhook https://my-function.com/webhooks/users \
  --filter "data.plan=premium"
```

## Webhook Security

### Signature Verification

Always verify webhook signatures:

1. Extract the signature from headers
2. Compute expected signature using secret
3. Use constant-time comparison
4. Reject if mismatch

### IP Allowlisting

Restrict webhook sources:

```yaml
webhook:
  allowed_ips:
    - 192.168.1.0/24
    - 10.0.0.0/8
```

### TLS Requirements

```yaml
webhook:
  require_tls: true  # Only accept HTTPS webhooks
  min_tls_version: "1.2"
```

## Best Practices

### Response Time

Respond quickly to webhooks:

```python
def handler(request):
    # Acknowledge immediately
    # Process asynchronously
    asyncio.create_task(process_async(request))
    
    return {
        'statusCode': 202,  # Accepted
        'body': json.dumps({'status': 'processing'})
    }
```

### Idempotency

Handle duplicate webhooks gracefully:

```python
def process_event(event):
    event_id = event['id']
    
    # Check if already processed
    if is_processed(event_id):
        return {'status': 'already_processed'}
    
    # Process and mark as done
    result = do_process(event)
    mark_processed(event_id)
    
    return result
```

### Error Handling

Return appropriate status codes:

- `200 OK`: Successfully processed
- `202 Accepted`: Processing asynchronously
- `400 Bad Request`: Invalid payload
- `401 Unauthorized`: Invalid signature
- `500 Internal Error`: Processing failed (will retry)

## Testing Webhooks

### Local Testing

```bash
# Use ngrok for local webhook testing
ngrok http 8080

# Configure webhook URL to ngrok URL
ffly webhook test \
  --url https://<ngrok-id>.ngrok.io/webhooks \
  --event function.completed
```

### Payload Inspection

```bash
# View recent webhook payloads
ffly webhook logs \
  --function webhook-handler \
  --limit 10
```

## Advanced Features

### Webhook Batching

Combine multiple events:

```yaml
webhook:
  batch:
    enabled: true
    max_size: 100
    max_wait: 5s
```

### Transformations

Transform payloads before sending:

```yaml
webhook:
  transform: |
    {
      "event": {{ event.type }},
      "data": {{ event.payload | json }},
      "timestamp": {{ now | iso8601 }}
    }
```

### Filtering

Filter events by criteria:

```bash
ffly events subscribe order.created \
  --webhook https://example.com/webhooks/orders \
  --filter "data.amount > 100" \
  --filter "data.currency = USD"
```

## Monitoring

### Webhook Logs

```bash
# View webhook delivery logs
ffly webhook logs \
  --function my-webhook \
  --status failed \
  --since 1h
```

### Metrics

- Delivery rate
- Response time
- Retry count
- Error rates by status code

## Troubleshooting

### Webhooks Not Received

- Verify URL is accessible
- Check firewall/allowlist settings
- Ensure function is deployed and running
- Review webhook logs for errors

### Signature Mismatch

- Verify secret is correctly configured
- Check signature algorithm (HMAC-SHA256)
- Ensure raw body is used for signature
- Check for encoding issues

### Timeouts

- Optimize handler to respond quickly
- Use async processing for long tasks
- Increase timeout if necessary
- Consider webhook batching

## Next Steps

- Set up [CI/CD integration](/guides/ci-cd/) for automated deployments
- Learn about [monitoring and analytics](/analytics/)
- Explore [rate limiting](/guides/rate-limiting/) to protect endpoints
