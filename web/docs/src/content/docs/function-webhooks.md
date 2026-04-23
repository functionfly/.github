---
title: "FUNCTION WEBHOOKS"
---

# Function Webhooks

Function Webhooks enable real-time notifications for function deployment events via HTTP callbacks.

## Overview

Subscribe to function events and receive signed HTTP POST requests when deployments occur, fail, scale, or are deleted.

## Event Types

| Event | Description |
|-------|-------------|
| `function.deployed` | Function deployed successfully |
| `function.failed` | Function deployment failed |
| `function.scaled` | Function scaled up or down |
| `function.deleted` | Function deleted |

## Model

### Subscription

```go
type FunctionWebhookSubscription struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    FunctionID  *uuid.UUID  // nil = all functions
    URL         string
    Secret      string      // HMAC signing secret
    EventTypes  []string
    Active      bool
    CreatedAt   time.Time
    CreatedBy   uuid.UUID
}
```

### Delivery

```go
type FunctionWebhookDelivery struct {
    ID             uuid.UUID
    SubscriptionID uuid.UUID
    EventType      string
    Payload        json.RawMessage
    ResponseStatus *int
    ResponseBody   *string
    AttemptedAt     time.Time
    Success        bool
    ErrorMessage   *string
}
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/function-webhooks` | Create subscription |
| GET | `/v1/function-webhooks` | List subscriptions |
| GET | `/v1/function-webhooks/{id}` | Get subscription |
| PUT | `/v1/function-webhooks/{id}` | Update subscription |
| DELETE | `/v1/function-webhooks/{id}` | Delete subscription |
| GET | `/v1/function-webhooks/{id}/deliveries` | List deliveries |
| POST | `/v1/function-webhooks/{id}/test` | Test webhook |

## Creating a Subscription

```bash
curl -X POST https://api.functionfly.com/v1/function-webhooks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "deployment-notifier",
    "url": "https://my-ci.example.com/webhook",
    "event_types": ["function.deployed", "function.failed"],
    "function_id": null
  }'
```

Response:
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "tenant_id": "...",
  "function_id": null,
  "url": "https://my-ci.example.com/webhook",
  "event_types": ["function.deployed", "function.failed"],
  "active": true,
  "created_at": "2026-04-21T18:57:00Z"
}
```

## Webhook Payload

```json
{
  "event_id": "evt_test_abc123...",
  "event_type": "function.deployed",
  "timestamp": "2026-04-21T18:58:00Z",
  "api_version": "2024-04-12",
  "test": false,
  "data": {
    "function_id": "...",
    "function_name": "my-function",
    "version": "v1.2.3",
    "tenant_id": "..."
  }
}
```

## Security

### HMAC Signature

Each webhook request includes signature headers for verification:

```
X-FunctionFly-Signature: sha256=<hmac>
X-FunctionFly-Timestamp: <unix_timestamp>
```

To verify, compute `HMAC-SHA256(secret, timestamp + payload)` and compare.

### URL Validation

- HTTPS required (localhost and private IPs blocked)
- Prevents SSRF attacks

## Delivery Behavior

- **Success**: HTTP 2xx response within 30s timeout
- **Failure**: Non-2xx response or timeout
- Deliveries are recorded regardless of success/failure
- No automatic retry (use test endpoint to verify)

## Database

### Table: `function_webhook_subscriptions`

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| tenant_id | UUID | Tenant ownership |
| function_id | UUID | Optional function filter |
| url | TEXT | Webhook endpoint URL |
| secret | VARCHAR(255) | HMAC signing secret |
| event_types | TEXT[] | Subscribed event types |
| active | BOOLEAN | Subscription active |
| created_at | TIMESTAMPTZ | Creation timestamp |
| created_by | UUID | User who created |

### Table: `function_webhook_deliveries`

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| subscription_id | UUID | Parent subscription |
| event_type | VARCHAR(50) | Event that triggered |
| payload | JSONB | Full webhook payload |
| response_status | INTEGER | HTTP status code |
| response_body | TEXT | Response body |
| attempted_at | TIMESTAMPTZ | Delivery attempt time |
| success | BOOLEAN | Delivery success |
| error_message | TEXT | Error details |

## Integration

In the deployment orchestrator, trigger events:

```go
functionWebhookService.TriggerFunctionEvent(
    ctx,
    "function.deployed",
    functionID,
    map[string]interface{}{
        "function_id":   functionID,
        "function_name": functionName,
        "version":       version,
    },
)
```