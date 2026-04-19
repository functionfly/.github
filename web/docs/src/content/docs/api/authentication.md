---
title: Authentication
description: FunctionFly API authentication methods
---

# Authentication

FunctionFly supports multiple authentication methods for API access.

## API Key Authentication

The most common method for programmatic access:

```http
Authorization: Bearer <your-api-key>
```

## OAuth Authentication

For user-based authentication:

```http
GET /v1/auth/oauth/providers
GET /v1/auth/oauth/url?provider=github&redirect_uri=...
```

## Session Authentication

For dashboard and web applications:

```http
POST /v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}
```

## Refresh Token

Refresh an expired access token:

```http
POST /v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "..."
}
```

## Rate Limits

| Endpoint Category | Limit |
|-------------------|-------|
| Authentication | 5 requests/minute |
| General API | 60 requests/minute |
