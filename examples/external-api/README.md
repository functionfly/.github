# External API Example

This example demonstrates how to use the `external_api` capability in FunctionFly functions.

## Overview

The external API capability allows functions to make HTTP requests to external services. This enables integration with third-party APIs, webhooks, and other internet services.

## Configuration

Set the following environment variables:

- `API_BASE_URL`: Base URL for API calls (default: https://httpbin.org)
- `API_TIMEOUT`: Request timeout in seconds (default: 10)

## Usage Examples

### GET Request

```json
{
  "method": "GET",
  "endpoint": "/get",
  "headers": {
    "Authorization": "Bearer token123"
  }
}
```

### POST Request

```json
{
  "method": "POST",
  "endpoint": "/post",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

## Response

```json
{
  "status": "success",
  "request": {
    "method": "POST",
    "url": "https://httpbin.org/post",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "name": "John Doe",
      "email": "john@example.com"
    }
  },
  "response": {
    "url": "https://httpbin.org/post",
    "method": "POST",
    "data": {
      "name": "John Doe",
      "email": "john@example.com"
    },
    "simulated": true
  },
  "response_length": 145
}
```

## Security Notes

- External API calls are rate-limited per function
- Only HTTPS URLs are allowed (HTTP may be blocked)
- Request/response size limits apply
- Certain domains may be blocked for security reasons