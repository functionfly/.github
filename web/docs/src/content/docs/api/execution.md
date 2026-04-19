---
title: Execution API
description: Execute functions via API
---

# Execution API

Execute functions and retrieve results via the API.

## Execute Function (Public)

For public functions that don't require authentication:

```http
POST /v1/execute/{functionId}
Content-Type: application/json

{
  "data": {
    "key": "value"
  }
}
```

## Execute Function (Authenticated)

For private functions requiring authentication:

```http
POST /v1/run/{functionId}
Authorization: Bearer <token>
Content-Type: application/json

{
  "input": {
    "key": "value"
  }
}
```

## Get Execution Result

```http
GET /v1/executions/{id}
```

## Response Format

```json
{
  "execution_id": "exec_123456",
  "status": "completed",
  "result": {
    "status": "ok",
    "data": {}
  },
  "duration_ms": 125,
  "executed_at": "2024-01-01T00:00:00Z"
}
```

## Execution Status

| Status | Description |
|--------|-------------|
| `pending` | Execution queued |
| `running` | Currently executing |
| `completed` | Successfully completed |
| `failed` | Execution failed |
| `timeout` | Execution timed out |

## Rate Limits

| Endpoint | Limit |
|----------|-------|
| Function Execution | 100 requests/minute |
