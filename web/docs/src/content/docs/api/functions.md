---
title: Functions API
description: Manage serverless functions via API
---

# Functions API

Create, deploy, and manage serverless functions programmatically.

## List Functions

```http
GET /v1/functions
```

## Get Function

```http
GET /v1/functions/{id}
```

## Create Function

```http
POST /v1/functions
Content-Type: application/json

{
  "name": "my-function",
  "runtime": "python",
  "code": "def main(request): return {'status': 'ok'}"
}
```

## Update Function

```http
PATCH /v1/functions/{id}
Content-Type: application/json

{
  "name": "updated-name",
  "code": "..."
}
```

## Delete Function

```http
DELETE /v1/functions/{id}
```

## Deploy Function

```http
POST /v1/functions/{id}/deploy
```

## Get Function Stats

```http
GET /v1/functions/{id}/stats
```

## Response Format

```json
{
  "id": "fn_123456",
  "name": "my-function",
  "runtime": "python",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z"
}
```
