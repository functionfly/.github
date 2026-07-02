---
title: Marketplace API
description: Full API reference for the FunctionFly Marketplace
sidebar:
  order: 4
---


---

## Unified Search

### Search

```
GET /v1/marketplace/search
```

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `q` | string | — | Search query |
| `type` | string | — | `agent`, `function`, `extension` (empty = all) |
| `limit` | int | 20 | Page size |
| `offset` | int | 0 | Pagination offset |

```json
{
  "items": [
    {
      "id": "ext_001",
      "type": "extension",
      "name": "GitHub Integration",
      "description": "Connect functions to GitHub repos",
      "rating": 4.8,
      "install_count": 1240
    }
  ],
  "total": 42
}
```

---

## Extensions

### List Extensions

```
GET /v1/marketplace/extensions
```

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `category` | string | — | Filter by category |
| `status` | string | — | `draft`, `published`, `active` |
| `search` | string | — | Text search |
| `featured` | bool | — | Featured only |
| `tags` | string | — | Comma-separated tags |
| `sort` | string | `trending` | `trending`, `top_rated`, `newest`, `most_installed` |
| `limit` | int | 20 | Page size |
| `offset` | int | 0 | Pagination offset |

### Create Extension

```
POST /v1/marketplace/extensions
```

### Get Extension

```
GET /v1/marketplace/extensions/{id}
```

### Update Extension

```
PUT /v1/marketplace/extensions/{id}
```

### Delete Extension

```
DELETE /v1/marketplace/extensions/{id}
```

### Install Extension

```
POST /v1/marketplace/extensions/{id}/install
```

Creates a plugin in the user's workspace.

### Rate Extension

```
POST /v1/marketplace/extensions/{id}/rate
```

```json
{
  "rating": 5,
  "review": "Excellent plugin, works perfectly"
}
```

### Get My Rating

```
GET /v1/marketplace/extensions/{id}/my-rating
```

### List Ratings

```
GET /v1/marketplace/extensions/{id}/ratings
```

### Check for Updates

```
POST /v1/marketplace/check-updates
```

### Get Install Counts

```
GET /v1/marketplace/install-counts
```

### List Categories

```
GET /v1/marketplace/categories
```

---

## Agent Ratings

### Rate Agent

```
POST /v1/marketplace/agents/{id}/rate
```

```json
{
  "rating": 4,
  "review": "Great code reviewer"
}
```

### Get My Agent Rating

```
GET /v1/marketplace/agents/{id}/my-rating
```

### List Agent Ratings

```
GET /v1/marketplace/agents/{id}/ratings
```

---

## Function Ratings

### Rate Function

```
POST /v1/marketplace/functions/{id}/rate
```

```json
{
  "rating": 5,
  "review": "Fast and accurate PDF summarizer"
}
```

### Get My Function Rating

```
GET /v1/marketplace/functions/{id}/my-rating
```

### List Function Ratings

```
GET /v1/marketplace/functions/{id}/ratings
```

---

## Error Codes

| Status | Code | Meaning |
|--------|------|---------|
| 400 | `INVALID_LISTING` | Missing required fields |
| 403 | `NOT_OWNER` | Only the creator can modify |
| 404 | `LISTING_NOT_FOUND` | Listing ID not found |
| 409 | `ALREADY_LISTED` | Asset already listed |
| 422 | `SECURITY_CHECK_FAILED` | Extension failed security analysis |
