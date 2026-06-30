---
title: REST API Reference
description: FunctionFly REST API overview
---

# REST API

The FunctionFly REST API provides programmatic access to all platform features.

## Base URL

```
https://api.functionfly.com/v1
```

## Authentication

Include your API key in the Authorization header:

```bash
curl -H "Authorization: Bearer $FFLY_API_KEY" \
  https://api.functionfly.com/v1/functions
```

## Endpoints

### Functions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/functions` | List all functions |
| POST | `/functions` | Create a function |
| GET | `/functions/:id` | Get function details |
| DELETE | `/functions/:id` | Delete a function |

### Execution

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/execute/:name` | Execute a function |
| GET | `/executions/:id` | Get execution result |

### Agents

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/agents` | List agents |
| POST | `/agents` | Create an agent |
| POST | `/agents/:id/execute` | Run agent |

## Rate Limits

| Plan | Requests/min |
|------|-------------|
| Free | 60 |
| Pro | 600 |
| Enterprise | 6000 |

## next steps

- [Authentication](/api/authentication/)
- [Functions API](/api/functions/)
- [Execution API](/api/execution/)
