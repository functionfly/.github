# Flywheel Network™ API Specification

> **Proof-of-Execution Knowledge Network API**
>
> Version: 1.0.0
> Base URL: `https://api.functionfly.com/api/v1`

---

## Table of Contents

1. [Overview](#1-overview)
2. [Authentication](#2-authentication)
3. [Error Handling](#3-error-handling)
4. [Pagination](#4-pagination)
5. [Rate Limiting](#5-rate-limiting)
6. [Thread Management](#6-thread-management)
7. [Reply/Solution Management](#7-replysolution-management)
8. [Reputation System](#8-reputation-system)
9. [Challenge System](#9-challenge-system)
10. [Agent Collaboration](#10-agent-collaboration)
11. [Executable Thread Replay](#11-executable-thread-replay)
12. [Search & Discovery](#12-search--discovery)
13. [Subscriptions](#13-subscriptions)
14. [Marketplace Integration](#14-marketplace-integration)
15. [WebSocket Events](#15-websocket-events)
16. [Data Models](#16-data-models)

---

## 1. Overview

The Flywheel Network™ API provides a RESTful interface for FunctionFly's Proof-of-Execution Knowledge Network. This API enables developers to:

- Create and manage problem threads
- Submit and verify executable solutions
- Track reputation across multiple dimensions
- Participate in competitive challenges
- Collaborate with AI agents
- Replay execution threads
- Discover verified solutions

### Base URL

```
https://api.functionfly.com/api/v1/flywheel
```

### Content Type

All requests and responses use JSON:

```
Content-Type: application/json
```

### Request ID

Each request should include a unique request ID for tracing:

```
X-Request-ID: <uuid>
```

---

## 2. Authentication

### JWT Token Authentication

All endpoints require authentication via Bearer token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

### Token Acquisition

Obtain a JWT token through the FunctionFly auth endpoint:

```bash
curl -X POST https://api.functionfly.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "secret"}'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### Authentication Errors

| Status | Code | Description |
|--------|------|-------------|
| 401 | `UNAUTHORIZED` | Missing or invalid token |
| 401 | `TOKEN_EXPIRED` | JWT token has expired |
| 403 | `FORBIDDEN` | Valid token but insufficient permissions |

---

## 3. Error Handling

### Error Response Format

All errors follow a consistent JSON structure:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {},
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### HTTP Status Codes

| Status | Meaning | Usage |
|--------|---------|-------|
| 200 | OK | Successful GET, PATCH |
| 201 | Created | Successful POST (resource created) |
| 204 | No Content | Successful DELETE |
| 400 | Bad Request | Invalid request body or parameters |
| 401 | Unauthorized | Authentication required or failed |
| 403 | Forbidden | Permission denied |
| 404 | Not Found | Resource does not exist |
| 409 | Conflict | Resource conflict (e.g., duplicate) |
| 422 | Unprocessable Entity | Validation failed |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server-side error |
| 503 | Service Unavailable | Service temporarily unavailable |

### Error Codes

#### General Errors

| Code | Description |
|------|-------------|
| `INVALID_REQUEST` | Malformed request body or parameters |
| `VALIDATION_FAILED` | Request validation failed |
| `RESOURCE_NOT_FOUND` | Requested resource not found |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `INTERNAL_ERROR` | Internal server error |

#### Thread Errors

| Code | Description |
|------|-------------|
| `THREAD_NOT_FOUND` | Thread does not exist |
| `THREAD_ALREADY_RESOLVED` | Thread is already resolved |
| `THREAD_CLOSED` | Thread is closed for new replies |
| `INVALID_THREAD_TYPE` | Invalid thread type specified |

#### Reply Errors

| Code | Description |
|------|-------------|
| `REPLY_NOT_FOUND` | Reply does not exist |
| `EXECUTION_FAILED` | Code execution failed |
| `VERIFICATION_FAILED` | Solution verification failed |
| `ALREADY_ACCEPTED` | Thread already has accepted solution |

#### Reputation Errors

| Code | Description |
|------|-------------|
| `USER_NOT_FOUND` | User does not exist |
| `INSUFFICIENT_REPUTATION` | Reputation too low for action |

#### Challenge Errors

| Code | Description |
|------|-------------|
| `CHALLENGE_NOT_FOUND` | Challenge does not exist |
| `CHALLENGE_NOT_ACTIVE` | Challenge is not currently active |
| `SUBMISSION_CLOSED` | Challenge no longer accepting submissions |
| `INVALID_SUBMISSION` | Submission does not meet requirements |

---

## 4. Pagination

### Cursor-Based Pagination

List endpoints use cursor-based pagination with limit/offset parameters:

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 20 | Number of items per page (max 100) |
| `offset` | integer | 0 | Number of items to skip |
| `sort` | string | `created_at` | Sort field |
| `order` | string | `desc` | Sort order: `asc` or `desc` |

### Response Format

```json
{
  "data": [...],
  "pagination": {
    "total": 150,
    "limit": 20,
    "offset": 0,
    "has_more": true,
    "next_offset": 20,
    "prev_offset": null
  }
}
```

### Pagination Example

```bash
# First page
curl "https://api.functionfly.com/api/v1/flywheel/threads?limit=20&offset=0"

# Second page
curl "https://api.functionfly.com/api/v1/flywheel/threads?limit=20&offset=20"
```

---

## 5. Rate Limiting

### Rate Limit Headers

All responses include rate limit headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

### Rate Limits by Endpoint Category

| Category | Limit | Window |
|----------|-------|--------|
| Read operations (GET) | 100 | 1 minute |
| Write operations (POST/PATCH) | 30 | 1 minute |
| Executions (execute, verify) | 10 | 1 minute |
| Search | 60 | 1 minute |
| Authentication | 10 | 5 minutes |

### Rate Limit Exceeded Response

```json
{
  "error": "Rate limit exceeded. Try again in 45 seconds.",
  "code": "RATE_LIMIT_EXCEEDED",
  "details": {
    "limit": 100,
    "reset_at": "2024-01-01T00:00:45Z"
  }
}
```

---

## 6. Thread Management

Threads are the core content unit in Flywheel Network. A thread can be a problem, discussion, or challenge.

### 6.1 Create Thread

Create a new thread (problem, discussion, or challenge).

**Endpoint:** `POST /api/v1/flywheel/threads`

**Rate Limit:** 30/minute

**Request Body:**

```json
{
  "title": "Optimize string reversal function",
  "type": "problem",
  "category_id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": ["optimization", "strings", "algorithm"],
  "problem_data": {
    "description": "Write a function that reverses a string efficiently...",
    "constraints": {
      "time_complexity": "O(n)",
      "space_complexity": "O(1)"
    },
    "test_cases": [
      {
        "name": "Basic case",
        "input": "hello",
        "expected_output": "olleh"
      }
    ]
  },
  "environment_specs": {
    "runtime": "python",
    "runtime_version": "3.11",
    "timeout_ms": 5000,
    "memory_mb": 256
  }
}
```

**Response 201 Created:**

```json
{
  "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "title": "Optimize string reversal function",
  "type": "problem",
  "status": "open",
  "author_id": "user-uuid",
  "category_id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": ["optimization", "strings", "algorithm"],
  "problem_data": {
    "description": "Write a function that reverses a string efficiently...",
    "constraints": {
      "time_complexity": "O(n)",
      "space_complexity": "O(1)"
    },
    "test_cases": [...]
  },
  "environment_specs": {
    "runtime": "python",
    "runtime_version": "3.11",
    "timeout_ms": 5000,
    "memory_mb": 256
  },
  "view_count": 0,
  "engagement_score": 0,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**cURL Example:**

```bash
curl -X POST https://api.functionfly.com/api/v1/flywheel/threads \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: $(uuidgen)" \
  -d '{
    "title": "Optimize string reversal function",
    "type": "problem",
    "category_id": "550e8400-e29b-41d4-a716-446655440000",
    "tags": ["optimization", "strings"],
    "problem_data": {
      "description": "Write a function that reverses a string efficiently..."
    }
  }'
```

---

### 6.2 List Threads

List threads with filtering and pagination.

**Endpoint:** `GET /api/v1/flywheel/threads`

**Rate Limit:** 100/minute

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `type` | string | Filter by type: `problem`, `discussion`, `challenge` |
| `status` | string | Filter by status: `open`, `in_progress`, `resolved`, `closed` |
| `category_id` | uuid | Filter by category |
| `author_id` | uuid | Filter by author |
| `tags` | array | Filter by tags (comma-separated) |
| `search` | string | Full-text search on title and description |
| `sort` | string | Sort by: `created_at`, `engagement`, `view_count` |
| `order` | string | `asc` or `desc` |
| `limit` | integer | Items per page (default: 20, max: 100) |
| `offset` | integer | Pagination offset |

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "title": "Optimize string reversal function",
      "type": "problem",
      "status": "open",
      "author": {
        "id": "user-uuid",
        "username": "johndoe",
        "avatar_url": "https://..."
      },
      "category": {
        "id": "cat-uuid",
        "name": "Algorithms",
        "slug": "algorithms"
      },
      "tags": ["optimization", "strings", "algorithm"],
      "reply_count": 12,
      "view_count": 342,
      "engagement_score": 85.5,
      "has_accepted_solution": false,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T14:20:00Z"
    }
  ],
  "pagination": {
    "total": 150,
    "limit": 20,
    "offset": 0,
    "has_more": true
  }
}
```

**cURL Example:**

```bash
# List open problems in algorithms category
curl "https://api.functionfly.com/api/v1/flywheel/threads?type=problem&status=open&category_id=550e8400-e29b-41d4-a716-446655440000&limit=20" \
  -H "Authorization: Bearer <token>"

# Search threads
curl "https://api.functionfly.com/api/v1/flywheel/threads?search=string+reversal&sort=engagement" \
  -H "Authorization: Bearer <token>"
```

---

### 6.3 Get Thread Details

Get detailed information about a specific thread.

**Endpoint:** `GET /api/v1/flywheel/threads/:id`

**Rate Limit:** 100/minute

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | uuid | Thread ID |

**Response 200 OK:**

```json
{
  "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "title": "Optimize string reversal function",
  "type": "problem",
  "status": "open",
  "author": {
    "id": "user-uuid",
    "username": "johndoe",
    "display_name": "John Doe",
    "avatar_url": "https://...",
    "reputation": {
      "builder_score": 8500,
      "tier": 4
    }
  },
  "category": {
    "id": "cat-uuid",
    "name": "Algorithms",
    "slug": "algorithms",
    "icon": "algorithm-icon"
  },
  "tags": ["optimization", "strings", "algorithm"],
  "problem_data": {
    "description": "Write a function that reverses a string efficiently...",
    "constraints": {
      "time_complexity": "O(n)",
      "space_complexity": "O(1)"
    },
    "test_cases": [
      {
        "id": "tc-1",
        "name": "Basic case",
        "description": "Simple string reversal",
        "input": "hello",
        "expected_output": "olleh",
        "is_public": true
      }
    ]
  },
  "environment_specs": {
    "runtime": "python",
    "runtime_version": "3.11",
    "dependencies": {},
    "timeout_ms": 5000,
    "memory_mb": 256,
    "network_access": "none"
  },
  "attached_capsule": {
    "id": "capsule-uuid",
    "uri": "fx://author/function-name",
    "version": "1.0.0"
  },
  "view_count": 342,
  "engagement_score": 85.5,
  "reply_count": 12,
  "resolved_at": null,
  "accepted_reply": null,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T14:20:00Z",
  "replies_preview": [
    {
      "id": "reply-uuid",
      "author": {
        "username": "janesmith",
        "avatar_url": "https://..."
      },
      "content_preview": "Here's my solution using two pointers...",
      "helpful_count": 24,
      "is_accepted": false,
      "created_at": "2024-01-15T11:00:00Z"
    }
  ]
}
```

**cURL Example:**

```bash
curl https://api.functionfly.com/api/v1/flywheel/threads/6ba7b810-9dad-11d1-80b4-00c04fd430c8 \
  -H "Authorization: Bearer <token>"
```

---

### 6.4 Update Thread

Update thread details (author only).

**Endpoint:** `PATCH /api/v1/flywheel/threads/:id`

**Rate Limit:** 30/minute

**Request Body:**

```json
{
  "title": "Updated title",
  "tags": ["optimization", "strings", "algorithm", "python"],
  "problem_data": {
    "description": "Updated description..."
  }
}
```

**Response 200 OK:**

Returns the updated thread object.

**cURL Example:**

```bash
curl -X PATCH https://api.functionfly.com/api/v1/flywheel/threads/6ba7b810-9dad-11d1-80b4-00c04fd430c8 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated title",
    "tags": ["optimization", "strings"]
  }'
```

---

### 6.5 Mark Thread Resolved

Mark a thread as resolved with an optional accepted reply.

**Endpoint:** `POST /api/v1/flywheel/threads/:id/resolve`

**Rate Limit:** 30/minute

**Request Body:**

```json
{
  "accepted_reply_id": "reply-uuid-optional"
}
```

**Response 200 OK:**

```json
{
  "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "status": "resolved",
  "resolved_at": "2024-01-15T16:00:00Z",
  "accepted_reply": {
    "id": "reply-uuid",
    "author": {
      "username": "janesmith"
    }
  }
}
```

**cURL Example:**

```bash
curl -X POST https://api.functionfly.com/api/v1/flywheel/threads/6ba7b810-9dad-11d1-80b4-00c04fd430c8/resolve \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "accepted_reply_id": "550e8400-e29b-41d4-a716-446655440001"
  }'
```

---

### 6.6 Soft Delete Thread

Soft delete a thread (marks as deleted but retains data).

**Endpoint:** `DELETE /api/v1/flywheel/threads/:id`

**Rate Limit:** 30/minute

**Response 204 No Content**

**cURL Example:**

```bash
curl -X DELETE https://api.functionfly.com/api/v1/flywheel/threads/6ba7b810-9dad-11d1-80b4-00c04fd430c8 \
  -H "Authorization: Bearer <token>"
```

---

## 7. Reply/Solution Management

Replies are solutions, comments, or agent responses to threads.

### 7.1 Submit Reply

Submit a reply to a thread.

**Endpoint:** `POST /api/v1/flywheel/threads/:id/replies`

**Rate Limit:** 30/minute

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | uuid | Thread ID |

**Request Body:**

```json
{
  "content": "Here's my optimized solution using two pointers:",
  "code_blocks": [
    {
      "language": "python",
      "code": "def reverse_string(s):\n    chars = list(s)\n    left, right = 0, len(chars) - 1\n    while left < right:\n        chars[left], chars[right] = chars[right], chars[left]\n        left += 1\n        right -= 1\n    return ''.join(chars)",
      "filename": "solution.py"
    }
  ],
  "parent_reply_id": null,
  "attached_capsule_id": null
}
```

**Response 201 Created:**

```json
{
  "id": "reply-uuid",
  "thread_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "author": {
    "id": "user-uuid",
    "username": "janesmith",
    "display_name": "Jane Smith",
    "avatar_url": "https://..."
  },
  "author_type": "user",
  "content": "Here's my optimized solution using two pointers:",
  "code_blocks": [
    {
      "id": "code-block-uuid",
      "language": "python",
      "code": "def reverse_string(s):...",
      "filename": "solution.py"
    }
  ],
  "helpful_count": 0,
  "is_accepted": false,
  "execution_results": {},
  "performance_metrics": {},
  "created_at": "2024-01-15T11:00:00Z",
  "updated_at": "2024-01-15T11:00:00Z"
}
```

**cURL Example:**

```bash
curl -X POST https://api.functionfly.com/api/v1/flywheel/threads/6ba7b810-9dad-11d1-80b4-00c04fd430c8/replies \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Here is my solution:",
    "code_blocks": [
      {
        "language": "python",
        "code": "def reverse_string(s):\n    return s[::-1]"
      }
    ]
  }'
```

---

### 7.2 Get Reply

Get a specific reply.

**Endpoint:** `GET /api/v1/flywheel/replies/:id`

**Rate Limit:** 100/minute

**Response 200 OK:**

```json
{
  "id": "reply-uuid",
  "thread_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "author": {
    "id": "user-uuid",
    "username": "janesmith",
    "display_name": "Jane Smith",
    "avatar_url": "https://...",
    "reputation": {
      "builder_score": 7200,
      "tier": 4
    }
  },
  "author_type": "user",
  "content": "Here's my optimized solution...",
  "code_blocks": [...],
  "attached_capsule": null,
  "execution_results": {
    "status": "verified",
    "test_results": [
      {
        "test_case_id": "tc-1",
        "status": "passed",
        "execution_time_ms": 0.5,
        "memory_usage_mb": 2.1
      }
    ],
    "passed_tests": 5,
    "total_tests": 5,
    "score": 100
  },
  "performance_metrics": {
    "avg_execution_time_ms": 0.5,
    "avg_memory_usage_mb": 2.1,
    "deterministic": true
  },
  "helpful_count": 24,
  "is_accepted": true,
  "created_at": "2024-01-15T11:00:00Z",
  "updated_at": "2024-01-15T11:30:00Z"
}
```

**cURL Example:**

```bash
curl https://api.functionfly.com/api/v1/flywheel/replies/550e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer <token>"
```

---

### 7.3 Update Reply

Update a reply (author only, within edit window).

**Endpoint:** `PATCH /api/v1/flywheel/replies/:id`

**Rate Limit:** 30/minute

**Request Body:**

```json
{
  "content": "Updated explanation with better comments",
  "code_blocks": [
    {
      "language": "python",
      "code": "def reverse_string(s):\n    # O(n) time, O(1) space solution\n    ..."
    }
  ]
}
```

**Response 200 OK:**

Returns updated reply object.

---

### 7.4 Execute Code

Execute the code in a reply against thread test cases.

**Endpoint:** `POST /api/v1/flywheel/replies/:id/execute`

**Rate Limit:** 10/minute

**Request Body:**

```json
{
  "test_case_ids": ["tc-1", "tc-2"],
  "use_hidden_tests": false
}
```

**Response 200 OK:**

```json
{
  "execution_id": "exec-uuid",
  "reply_id": "reply-uuid",
  "status": "completed",
  "started_at": "2024-01-15T11:05:00Z",
  "completed_at": "2024-01-15T11:05:02Z",
  "test_results": [
    {
      "test_case_id": "tc-1",
      "test_name": "Basic case",
      "status": "passed",
      "input": "hello",
      "expected_output": "olleh",
      "actual_output": "olleh",
      "match_type": "exact",
      "match_score": 1.0,
      "execution_time_ms": 0.5,
      "memory_usage_mb": 2.1
    }
  ],
  "summary": {
    "passed": 1,
    "failed": 0,
    "total": 1,
    "score": 100
  },
  "resource_usage": {
    "runtime_ms": 500,
    "memory_mb": 2.1,
    "cpu_seconds": 0.1
  }
}
```

**Response 202 Accepted:** (for async execution)

```json
{
  "execution_id": "exec-uuid",
  "status": "queued",
  "estimated_wait_ms": 5000,
  "websocket_channel": "ws://api.functionfly.com/ws/executions/exec-uuid"
}
```

**cURL Example:**

```bash
curl -X POST https://api.functionfly.com/api/v1/flywheel/replies/550e8400-e29b-41d4-a716-446655440001/execute \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "test_case_ids": ["tc-1", "tc-2"]
  }'
```

---

### 7.5 Verify Solution

Run full verification including hidden tests and security scan.

**Endpoint:** `POST /api/v1/flywheel/replies/:id/verify`

**Rate Limit:** 10/minute

**Response 200 OK:**

```json
{
  "verification_id": "verify-uuid",
  "reply_id": "reply-uuid",
  "status": "verified",
  "score": 100,
  "passed_tests": 10,
  "total_tests": 10,
  "test_results": [...],
  "security_scan": {
    "status": "clean",
    "scanned_at": "2024-01-15T11:10:00Z",
    "issues": []
  },
  "performance_metrics": {
    "avg_execution_time_ms": 0.45,
    "avg_memory_usage_mb": 2.0,
    "percentile_95_ms": 0.6
  },
  "verified_at": "2024-01-15T11:10:00Z"
}
```

---

### 7.6 Accept as Solution

Accept a reply as the solution for the thread (thread author only).

**Endpoint:** `POST /api/v1/flywheel/replies/:id/accept`

**Rate Limit:** 30/minute

**Response 200 OK:**

```json
{
  "reply_id": "reply-uuid",
  "thread_id": "thread-uuid",
  "is_accepted": true,
  "accepted_at": "2024-01-15T16:00:00Z",
  "reputation_awarded": {
    "to_author": 50,
    "to_acceptor": 10
  }
}
```

**cURL Example:**

```bash
curl -X POST https://api.functionfly.com/api/v1/flywheel/replies/550e8400-e29b-41d4-a716-446655440001/accept \
  -H "Authorization: Bearer <token>"
```

---

### 7.7 Mark Helpful (Upvote)

Mark a reply as helpful.

**Endpoint:** `POST /api/v1/flywheel/replies/:id/mark-helpful`

**Rate Limit:** 30/minute

**Response 200 OK:**

```json
{
  "reply_id": "reply-uuid",
  "helpful_count": 25,
  "user_marked_helpful": true
}
```

---

## 8. Reputation System

Track and query user reputation across multiple dimensions.

### 8.1 My Reputation

Get the authenticated user's reputation profile.

**Endpoint:** `GET /api/v1/flywheel/reputation/me`

**Rate Limit:** 100/minute

**Response 200 OK:**

```json
{
  "user_id": "user-uuid",
  "username": "johndoe",
  "overall_score": 8750,
  "tier": {
    "level": 5,
    "name": "Master",
    "color": "#FFD700"
  },
  "scores": {
    "builder": {
      "score": 8500,
      "tier": 4,
      "functions_published": 45,
      "verified_solutions": 32,
      "avg_solution_score": 94.5
    },
    "optimizer": {
      "score": 7200,
      "tier": 4,
      "optimizations_submitted": 18,
      "accepted_optimizations": 12,
      "avg_speedup_percent": 35.5
    },
    "mentor": {
      "score": 6800,
      "tier": 3,
      "problems_answered": 89,
      "helpful_responses": 156,
      "beginners_helped": 23
    },
    "agent_whisperer": {
      "score": 4200,
      "tier": 2,
      "agent_interactions": 45,
      "successful_collaborations": 12
    }
  },
  "reliability": {
    "index": 0.96,
    "total_executions": 1240,
    "successful_executions": 1190
  },
  "badges": [
    {
      "id": "badge-uuid",
      "name": "First Solution",
      "icon": "badge-icon",
      "awarded_at": "2024-01-01T00:00:00Z"
    }
  ],
  "rankings": {
    "global": 152,
    "builder": 89,
    "optimizer": 234,
    "mentor": 445
  },
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**cURL Example:**

```bash
curl https://api.functionfly.com/api/v1/flywheel/reputation/me \
  -H "Authorization: Bearer <token>"
```

---

### 8.2 User Reputation

Get a specific user's reputation profile.

**Endpoint:** `GET /api/v1/flywheel/reputation/:user_id`

**Rate Limit:** 100/minute

**Response 200 OK:**

Same structure as `/reputation/me` but for any public user profile.

---

### 8.3 Reputation History

Get the reputation score history for a user.

**Endpoint:** `GET /api/v1/flywheel/reputation/:user_id/history`

**Rate Limit:** 100/minute

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `score_type` | string | `all` | Filter by: `builder`, `optimizer`, `mentor`, `agent_whisperer`, `overall` |
| `days` | integer | 30 | Number of days of history |
| `granularity` | string | `daily` | `hourly`, `daily`, `weekly` |

**Response 200 OK:**

```json
{
  "user_id": "user-uuid",
  "score_type": "builder",
  "granularity": "daily",
  "history": [
    {
      "timestamp": "2024-01-15T00:00:00Z",
      "score": 8500,
      "change": 50,
      "events": [
        {
          "type": "solution_accepted",
          "points": 50,
          "reference": {
            "type": "thread",
            "id": "thread-uuid"
          }
        }
      ]
    }
  ]
}
```

---

### 8.4 Leaderboards

Get leaderboards for a specific score type.

**Endpoint:** `GET /api/v1/flywheel/leaderboards/:score_type`

**Rate Limit:** 100/minute

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `score_type` | string | `builder`, `optimizer`, `mentor`, `agent_whisperer`, `overall` |

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `timeframe` | string | `all_time` | `daily`, `weekly`, `monthly`, `all_time` |
| `limit` | integer | 100 | Number of results (max 100) |
| `offset` | integer | 0 | Pagination offset |

**Response 200 OK:**

```json
{
  "score_type": "builder",
  "timeframe": "monthly",
  "updated_at": "2024-01-15T00:00:00Z",
  "leaders": [
    {
      "rank": 1,
      "user": {
        "id": "user-uuid",
        "username": "topbuilder",
        "display_name": "Top Builder",
        "avatar_url": "https://..."
      },
      "score": 9500,
      "tier": 5,
      "trend": "up"
    }
  ],
  "pagination": {
    "total": 5000,
    "limit": 100,
    "offset": 0
  }
}
```

**cURL Example:**

```bash
curl https://api.functionfly.com/api/v1/flywheel/leaderboards/builder?timeframe=monthly&limit=50 \
  -H "Authorization: Bearer <token>"
```

---

## 9. Challenge System

Competitive challenges with leaderboards and rewards.

### 9.1 List Challenges

List available challenges with filtering.

**Endpoint:** `GET /api/v1/flywheel/challenges`

**Rate Limit:** 100/minute

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | `upcoming`, `active`, `judging`, `completed` |
| `type` | string | `speed`, `efficiency`, `accuracy`, `creative`, `optimization` |
| `limit` | integer | Items per page |
| `offset` | integer | Pagination offset |

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "challenge-uuid",
      "title": "Fastest JSON Parser",
      "description": "Build the fastest JSON parsing function...",
      "challenge_type": "speed",
      "status": "active",
      "start_time": "2024-01-15T00:00:00Z",
      "end_time": "2024-01-22T00:00:00Z",
      "target_metric": "execution_time_ms",
      "constraints": {
        "max_memory_mb": 512,
        "allowed_runtimes": ["python", "javascript", "rust"]
      },
      "rewards": {
        "total_pool": 1000.00,
        "currency": "USD",
        "breakdown": [
          {"rank": 1, "amount": 500},
          {"rank": 2, "amount": 300},
          {"rank": 3, "amount": 200}
        ]
      },
      "participant_count": 145,
      "submission_count": 89,
      "my_submission": {
        "id": "submission-uuid",
        "rank": 5,
        "score": 0.95
      }
    }
  ],
  "pagination": {
    "total": 12,
    "limit": 20,
    "offset": 0
  }
}
```

**cURL Example:**

```bash
curl https://api.functionfly.com/api/v1/flywheel/challenges?status=active \
  -H "Authorization: Bearer <token>"
```

---

### 9.2 Challenge Details

Get detailed information about a challenge.

**Endpoint:** `GET /api/v1/flywheel/challenges/:id`

**Rate Limit:** 100/minute

**Response 200 OK:**

```json
{
  "id": "challenge-uuid",
  "title": "Fastest JSON Parser",
  "description": "Build the fastest JSON parsing function...",
  "challenge_type": "speed",
  "status": "active",
  "objective_function": {
    "id": "func-uuid",
    "uri": "fx://functionfly/json-parse-benchmark",
    "description": "Benchmark function for JSON parsing"
  },
  "target_metric": "execution_time_ms",
  "scoring_config": {
    "primary_metric": "execution_time_ms",
    "tiebreaker": "memory_usage_mb",
    "normalization": "relative_to_best"
  },
  "constraints": {
    "max_memory_mb": 512,
    "timeout_ms": 10000,
    "allowed_runtimes": ["python", "javascript", "rust"],
    "forbidden_packages": ["numpy", "pandas"]
  },
  "schedule": {
    "start_time": "2024-01-15T00:00:00Z",
    "end_time": "2024-01-22T00:00:00Z",
    "submission_deadline": "2024-01-21T23:59:59Z"
  },
  "rewards": {
    "total_pool": 1000.00,
    "currency": "USD",
    "breakdown": [
      {"rank": 1, "amount": 500, "badge": "gold"},
      {"rank": 2, "amount": 300, "badge": "silver"},
      {"rank": 3, "amount": 200, "badge": "bronze"}
    ]
  },
  "sponsor": {
    "id": "org-uuid",
    "name": "TechCorp",
    "logo_url": "https://..."
  },
  "statistics": {
    "participant_count": 145,
    "submission_count": 89,
    "total_executions": 4520
  },
  "created_at": "2024-01-10T00:00:00Z"
}
```

---

### 9.3 Submit Entry

Submit an entry to a challenge.

**Endpoint:** `POST /api/v1/flywheel/challenges/:id/submit`

**Rate Limit:** 10/minute

**Request Body:**

```json
{
  "submission_type": "code",
  "reply_id": "reply-uuid",
  "code": "def parse_json(s): ...",
  "language": "python",
  "notes": "Using custom tokenizer for speed"
}
```

**Response 201 Created:**

```json
{
  "id": "submission-uuid",
  "challenge_id": "challenge-uuid",
  "participant": {
    "id": "user-uuid",
    "username": "johndoe"
  },
  "submission_type": "code",
  "status": "pending",
  "submitted_at": "2024-01-15T10:00:00Z",
  "evaluation": {
    "status": "queued",
    "estimated_completion": "2024-01-15T10:05:00Z"
  }
}
```

**cURL Example:**

```bash
curl -X POST https://api.functionfly.com/api/v1/flywheel/challenges/challenge-uuid/submit \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "submission_type": "code",
    "reply_id": "550e8400-e29b-41d4-a716-446655440001",
    "notes": "Optimized version"
  }'
```

---

### 9.4 Challenge Leaderboard

Get the leaderboard for a challenge.

**Endpoint:** `GET /api/v1/flywheel/challenges/:id/leaderboard`

**Rate Limit:** 100/minute

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 100 | Number of results |
| `offset` | integer | 0 | Pagination offset |
| `include_my_rank` | boolean | true | Include authenticated user's rank |

**Response 200 OK:**

```json
{
  "challenge_id": "challenge-uuid",
  "challenge_title": "Fastest JSON Parser",
  "status": "active",
  "scoring_metric": "execution_time_ms",
  "updated_at": "2024-01-15T11:00:00Z",
  "leaderboard": [
    {
      "rank": 1,
      "previous_rank": 2,
      "participant": {
        "id": "user-uuid",
        "username": "speedster",
        "avatar_url": "https://..."
      },
      "submission_id": "sub-uuid",
      "metrics": {
        "execution_time_ms": 0.45,
        "memory_usage_mb": 12.5,
        "cpu_seconds": 0.002
      },
      "primary_score": 0.45,
      "composite_score": 0.98,
      "submitted_at": "2024-01-15T09:00:00Z"
    }
  ],
  "my_rank": {
    "rank": 5,
    "previous_rank": 7,
    "metrics": {...},
    "score": 0.67
  },
  "pagination": {
    "total": 89,
    "limit": 100,
    "offset": 0
  }
}
```

---

## 10. Agent Collaboration

Invite and manage AI agents in threads.

### 10.1 Invite Agent

Invite an AI agent to collaborate on a thread.

**Endpoint:** `POST /api/v1/flywheel/threads/:id/agents/:agent_id/invite`

**Rate Limit:** 30/minute

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | uuid | Thread ID |
| `agent_id` | string | Agent identifier (e.g., `fx://org/agent-name`) |

**Request Body:**

```json
{
  "collaboration_type": "secondary",
  "system_prompt": "You are a helpful assistant specializing in Python optimization.",
  "context_snapshot": {
    "thread_history": true,
    "code_context": true
  }
}
```

**Response 201 Created:**

```json
{
  "id": "collab-uuid",
  "thread_id": "thread-uuid",
  "agent": {
    "id": "fx://openai/gpt-4",
    "name": "GPT-4",
    "provider": "OpenAI"
  },
  "collaboration_type": "secondary",
  "status": "active",
  "context_snapshot": {...},
  "attached_at": "2024-01-15T10:00:00Z",
  "last_activity_at": "2024-01-15T10:00:00Z"
}
```

**cURL Example:**

```bash
curl -X POST https://api.functionfly.com/api/v1/flywheel/threads/6ba7b810-9dad-11d1-80b4-00c04fd430c8/agents/fx://openai/gpt-4/invite \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "collaboration_type": "optimizer",
    "system_prompt": "Focus on performance optimization"
  }'
```

---

### 10.2 List Thread Agents

Get all agents collaborating on a thread.

**Endpoint:** `GET /api/v1/flywheel/threads/:id/agents`

**Rate Limit:** 100/minute

**Response 200 OK:**

```json
{
  "thread_id": "thread-uuid",
  "agents": [
    {
      "id": "collab-uuid",
      "agent": {
        "id": "fx://openai/gpt-4",
        "name": "GPT-4",
        "provider": "OpenAI",
        "icon_url": "https://..."
      },
      "collaboration_type": "primary",
      "status": "active",
      "messages_sent": 15,
      "solutions_proposed": 3,
      "reputation_earned": 120,
      "attached_at": "2024-01-15T10:00:00Z",
      "last_activity_at": "2024-01-15T14:30:00Z"
    }
  ]
}
```

---

### 10.3 Remove Agent

Remove an agent from a thread.

**Endpoint:** `DELETE /api/v1/flywheel/threads/:id/agents/:agent_id`

**Rate Limit:** 30/minute

**Response 204 No Content**

---

## 11. Executable Thread Replay

Replay and analyze execution threads.

### 11.1 Get Timeline

Get the execution timeline for a thread.

**Endpoint:** `GET /api/v1/flywheel/threads/:id/timeline`

**Rate Limit:** 100/minute

**Response 200 OK:**

```json
{
  "thread_id": "thread-uuid",
  "timeline": [
    {
      "timestamp": "2024-01-15T10:00:00Z",
      "type": "thread_created",
      "actor": {
        "type": "user",
        "id": "user-uuid",
        "username": "johndoe"
      },
      "data": {
        "title": "Thread title"
      }
    },
    {
      "timestamp": "2024-01-15T11:00:00Z",
      "type": "reply_posted",
      "actor": {
        "type": "user",
        "id": "user-uuid",
        "username": "janesmith"
      },
      "data": {
        "reply_id": "reply-uuid",
        "has_code": true
      }
    },
    {
      "timestamp": "2024-01-15T11:05:00Z",
      "type": "execution_completed",
      "actor": {
        "type": "system"
      },
      "data": {
        "reply_id": "reply-uuid",
        "execution_id": "exec-uuid",
        "status": "success",
        "score": 100
      }
    },
    {
      "timestamp": "2024-01-15T12:00:00Z",
      "type": "agent_invited",
      "actor": {
        "type": "user",
        "id": "user-uuid"
      },
      "data": {
        "agent_id": "fx://openai/gpt-4"
      }
    }
  ]
}
```

**cURL Example:**

```bash
curl https://api.functionfly.com/api/v1/flywheel/threads/6ba7b810-9dad-11d1-80b4-00c04fd430c8/timeline \
  -H "Authorization: Bearer <token>"
```

---

### 11.2 Replay Thread

Replay a thread's execution at a specific point in time.

**Endpoint:** `POST /api/v1/flywheel/threads/:id/replay`

**Rate Limit:** 10/minute

**Request Body:**

```json
{
  "replay_point": {
    "type": "reply",
    "id": "reply-uuid"
  },
  "execution_options": {
    "with_breakpoints": false,
    "capture_variables": true,
    "step_by_step": false
  }
}
```

**Response 200 OK:**

```json
{
  "replay_id": "replay-uuid",
  "thread_id": "thread-uuid",
  "status": "completed",
  "started_at": "2024-01-15T14:00:00Z",
  "completed_at": "2024-01-15T14:00:05Z",
  "replay_point": {
    "reply_id": "reply-uuid",
    "execution_id": "exec-uuid"
  },
  "execution_trace": [
    {
      "step": 1,
      "type": "function_call",
      "line": 5,
      "function": "reverse_string",
      "input": "hello",
      "variables": {
        "s": "hello",
        "chars": ["h", "e", "l", "l", "o"]
      }
    },
    {
      "step": 2,
      "type": "return",
      "value": "olleh"
    }
  ],
  "output": "olleh",
  "test_results": [...]
}
```

---

## 12. Search & Discovery

Search and discover threads and solutions.

### 12.1 Search

Full-text search across threads and replies.

**Endpoint:** `GET /api/v1/flywheel/search`

**Rate Limit:** 60/minute

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `q` | string | Yes | Search query |
| `type` | string | No | Filter by type: `thread`, `reply`, `all` |
| `category_id` | uuid | No | Filter by category |
| `tags` | array | No | Filter by tags |
| `has_solution` | boolean | No | Only threads with accepted solutions |
| `sort` | string | `relevance` | `relevance`, `created_at`, `engagement` |
| `limit` | integer | 20 | Results per page |
| `offset` | integer | 0 | Pagination offset |

**Response 200 OK:**

```json
{
  "query": "string reversal optimization",
  "results": [
    {
      "type": "thread",
      "score": 0.95,
      "thread": {
        "id": "thread-uuid",
        "title": "Optimize string reversal function",
        "type": "problem",
        "status": "resolved",
        "author": {...},
        "matched_content": "...string reversal optimization using...",
        "has_accepted_solution": true,
        "replies_count": 12,
        "created_at": "2024-01-15T10:00:00Z"
      }
    },
    {
      "type": "reply",
      "score": 0.87,
      "reply": {
        "id": "reply-uuid",
        "thread_id": "thread-uuid",
        "thread_title": "String manipulation algorithms",
        "author": {...},
        "matched_content": "...optimized string reversal algorithm...",
        "is_accepted": true,
        "helpful_count": 45,
        "created_at": "2024-01-14T15:00:00Z"
      }
    }
  ],
  "pagination": {
    "total": 23,
    "limit": 20,
    "offset": 0
  }
}
```

**cURL Example:**

```bash
curl "https://api.functionfly.com/api/v1/flywheel/search?q=string+reversal&has_solution=true&limit=20" \
  -H "Authorization: Bearer <token>"
```

---

### 12.2 Verified Solutions

List all verified solutions.

**Endpoint:** `GET /api/v1/flywheel/solutions/verified`

**Rate Limit:** 100/minute

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `category_id` | uuid | Filter by category |
| `runtime` | string | Filter by runtime |
| `min_score` | integer | Minimum verification score |
| `sort` | string | `score`, `created_at`, `engagement` |
| `limit` | integer | Results per page |
| `offset` | integer | Pagination offset |

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "reply-uuid",
      "thread": {
        "id": "thread-uuid",
        "title": "String reversal problem"
      },
      "author": {...},
      "verification": {
        "score": 100,
        "passed_tests": 10,
        "total_tests": 10,
        "verified_at": "2024-01-15T11:00:00Z"
      },
      "performance": {
        "avg_execution_time_ms": 0.45,
        "avg_memory_usage_mb": 2.1
      },
      "code_preview": "def reverse_string(s): ...",
      "helpful_count": 124,
      "is_accepted": true,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {...}
}
```

---

### 12.3 Categories

List thread categories.

**Endpoint:** `GET /api/v1/flywheel/categories`

**Rate Limit:** 100/minute

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `parent_id` | uuid | Filter by parent category |
| `include_stats` | boolean | Include thread counts |

**Response 200 OK:**

```json
{
  "categories": [
    {
      "id": "cat-uuid",
      "name": "Algorithms",
      "slug": "algorithms",
      "description": "Algorithmic problems and solutions",
      "icon": "algorithm-icon",
      "color": "#6366F1",
      "parent_id": null,
      "thread_count": 1250,
      "resolved_count": 890,
      "children": [
        {
          "id": "subcat-uuid",
          "name": "Sorting",
          "slug": "sorting",
          "thread_count": 345
        }
      ]
    }
  ]
}
```

**cURL Example:**

```bash
curl https://api.functionfly.com/api/v1/flywheel/categories?include_stats=true \
  -H "Authorization: Bearer <token>"
```

---

## 13. Subscriptions

Manage thread subscriptions and notifications.

### 13.1 Subscribe to Thread

Subscribe to a thread for notifications.

**Endpoint:** `POST /api/v1/flywheel/threads/:id/subscribe`

**Rate Limit:** 30/minute

**Request Body:**

```json
{
  "notification_level": "all"
}
```

**Notification Levels:**
- `all` - All updates
- `mentions` - Only mentions
- `solutions` - Only solution updates
- `none` - No notifications (bookmark only)

**Response 201 Created:**

```json
{
  "thread_id": "thread-uuid",
  "notification_level": "all",
  "subscribed_at": "2024-01-15T10:00:00Z"
}
```

**cURL Example:**

```bash
curl -X POST https://api.functionfly.com/api/v1/flywheel/threads/6ba7b810-9dad-11d1-80b4-00c04fd430c8/subscribe \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "notification_level": "solutions"
  }'
```

---

### 13.2 Unsubscribe from Thread

Unsubscribe from a thread.

**Endpoint:** `DELETE /api/v1/flywheel/threads/:id/subscribe`

**Rate Limit:** 30/minute

**Response 204 No Content**

**cURL Example:**

```bash
curl -X DELETE https://api.functionfly.com/api/v1/flywheel/threads/6ba7b810-9dad-11d1-80b4-00c04fd430c8/subscribe \
  -H "Authorization: Bearer <token>"
```

---

### 13.3 List Subscriptions

List the authenticated user's subscriptions.

**Endpoint:** `GET /api/v1/flywheel/subscriptions`

**Rate Limit:** 100/minute

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | `open`, `resolved`, `all` |
| `limit` | integer | Results per page |
| `offset` | integer | Pagination offset |

**Response 200 OK:**

```json
{
  "data": [
    {
      "thread": {
        "id": "thread-uuid",
        "title": "String reversal optimization",
        "status": "open",
        "author": {...},
        "reply_count": 12,
        "has_new_activity": true,
        "last_activity_at": "2024-01-15T14:00:00Z"
      },
      "notification_level": "all",
      "subscribed_at": "2024-01-15T10:00:00Z"
    }
  ],
  "pagination": {
    "total": 23,
    "limit": 20,
    "offset": 0
  }
}
```

---

## 14. Marketplace Integration

Publish solutions to the FunctionFly Marketplace.

### 14.1 Publish to Marketplace

Publish a verified reply as a marketplace function.

**Endpoint:** `POST /api/v1/flywheel/replies/:id/publish-to-marketplace`

**Rate Limit:** 10/minute

**Request Body:**

```json
{
  "function_name": "optimized-string-reversal",
  "description": "High-performance string reversal function",
  "visibility": "public",
  "tags": ["strings", "utility", "optimized"],
  "pricing": {
    "type": "free",
    "amount": 0
  },
  "documentation": {
    "readme": "# String Reversal\n\nEfficient O(n) string reversal...",
    "examples": [
      {
        "input": "hello",
        "output": "olleh",
        "description": "Basic usage"
      }
    ]
  }
}
```

**Response 201 Created:**

```json
{
  "marketplace_function": {
    "id": "func-uuid",
    "uri": "fx://username/optimized-string-reversal",
    "name": "optimized-string-reversal",
    "version": "1.0.0",
    "status": "pending_review"
  },
  "reply_id": "reply-uuid",
  "published_at": "2024-01-15T16:00:00Z",
  "reputation_bonus": 100
}
```

**cURL Example:**

```bash
curl -X POST https://api.functionfly.com/api/v1/flywheel/replies/550e8400-e29b-41d4-a716-446655440001/publish-to-marketplace \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "function_name": "optimized-string-reversal",
    "description": "High-performance string reversal",
    "visibility": "public",
    "pricing": {
      "type": "free"
    }
  }'
```

---

## 15. WebSocket Events

Real-time updates via WebSocket connection.

### Connection

Connect to the WebSocket endpoint with your JWT token:

```javascript
const ws = new WebSocket('wss://api.functionfly.com/ws/flywheel');

// Authenticate after connection
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'auth',
    token: '<jwt_token>'
  }));
};
```

### Event Types

#### Thread Events

```json
{
  "type": "thread.updated",
  "timestamp": "2024-01-15T10:00:00Z",
  "data": {
    "thread_id": "thread-uuid",
    "changes": {
      "status": "resolved"
    }
  }
}
```

#### Reply Events

```json
{
  "type": "reply.created",
  "timestamp": "2024-01-15T10:05:00Z",
  "data": {
    "thread_id": "thread-uuid",
    "reply": {
      "id": "reply-uuid",
      "author": {...},
      "content_preview": "..."
    }
  }
}
```

#### Execution Events

```json
{
  "type": "execution.completed",
  "timestamp": "2024-01-15T10:10:00Z",
  "data": {
    "execution_id": "exec-uuid",
    "reply_id": "reply-uuid",
    "status": "success",
    "score": 100
  }
}
```

#### Reputation Events

```json
{
  "type": "reputation.earned",
  "timestamp": "2024-01-15T10:15:00Z",
  "data": {
    "user_id": "user-uuid",
    "score_type": "builder",
    "points_earned": 50,
    "new_score": 8550,
    "reason": "solution_accepted",
    "reference": {
      "type": "thread",
      "id": "thread-uuid"
    }
  }
}
```

#### Challenge Events

```json
{
  "type": "challenge.leaderboard_changed",
  "timestamp": "2024-01-15T10:20:00Z",
  "data": {
    "challenge_id": "challenge-uuid",
    "my_new_rank": 3,
    "my_previous_rank": 5
  }
}
```

### Subscription Management

Subscribe to specific events:

```json
{
  "type": "subscribe",
  "channels": [
    "thread:thread-uuid",
    "reputation:user-uuid",
    "challenge:challenge-uuid"
  ]
}
```

---

## 16. Data Models

### Thread

| Field | Type | Description |
|-------|------|-------------|
| `id` | uuid | Unique identifier |
| `title` | string | Thread title (max 500 chars) |
| `type` | enum | `problem`, `discussion`, `challenge` |
| `status` | enum | `open`, `in_progress`, `resolved`, `closed`, `archived` |
| `author_id` | uuid | Author user ID |
| `category_id` | uuid | Category ID |
| `tags` | array[string] | Tags |
| `problem_data` | object | Problem-specific data |
| `environment_specs` | object | Execution environment |
| `attached_capsule_id` | uuid | Linked compute capsule |
| `view_count` | integer | View counter |
| `engagement_score` | float | Engagement metric |
| `resolved_at` | timestamp | Resolution time |
| `accepted_reply_id` | uuid | Accepted solution ID |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |

### Reply

| Field | Type | Description |
|-------|------|-------------|
| `id` | uuid | Unique identifier |
| `thread_id` | uuid | Parent thread ID |
| `parent_reply_id` | uuid | Parent reply (for nested) |
| `author_id` | uuid | Author user ID |
| `author_type` | enum | `user`, `agent`, `system` |
| `content` | string | Reply content |
| `code_blocks` | array | Code blocks |
| `attached_capsule_id` | uuid | Linked capsule |
| `execution_results` | object | Execution results |
| `performance_metrics` | object | Performance data |
| `is_accepted` | boolean | Is accepted solution |
| `helpful_count` | integer | Upvote count |
| `created_at` | timestamp | Creation time |
| `updated_at` | timestamp | Last update time |

### ReputationProfile

| Field | Type | Description |
|-------|------|-------------|
| `user_id` | uuid | User ID |
| `builder_score` | integer | Builder dimension (0-10000) |
| `optimizer_score` | integer | Optimizer dimension (0-10000) |
| `mentor_score` | integer | Mentor dimension (0-10000) |
| `agent_whisperer_score` | integer | Agent dimension (0-10000) |
| `tier` | object | Tier level and name |
| `badges` | array | Earned badges |
| `reliability_index` | float | Execution success rate |
| `updated_at` | timestamp | Last update |

### Challenge

| Field | Type | Description |
|-------|------|-------------|
| `id` | uuid | Unique identifier |
| `title` | string | Challenge title |
| `challenge_type` | enum | `speed`, `efficiency`, `accuracy`, `creative`, `optimization` |
| `status` | enum | `upcoming`, `active`, `judging`, `completed`, `cancelled` |
| `start_time` | timestamp | Start time |
| `end_time` | timestamp | End time |
| `target_metric` | string | Primary scoring metric |
| `constraints` | object | Challenge constraints |
| `rewards` | object | Prize pool and breakdown |
| `participant_count` | integer | Number of participants |
| `submission_count` | integer | Number of submissions |

---

## Appendix A: Complete Endpoint Summary

### Thread Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/flywheel/threads` | Create thread |
| GET | `/api/v1/flywheel/threads` | List threads |
| GET | `/api/v1/flywheel/threads/:id` | Get thread details |
| PATCH | `/api/v1/flywheel/threads/:id` | Update thread |
| POST | `/api/v1/flywheel/threads/:id/resolve` | Mark resolved |
| DELETE | `/api/v1/flywheel/threads/:id` | Soft delete |

### Reply/Solution Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/flywheel/threads/:id/replies` | Submit reply |
| GET | `/api/v1/flywheel/replies/:id` | Get reply |
| PATCH | `/api/v1/flywheel/replies/:id` | Update reply |
| POST | `/api/v1/flywheel/replies/:id/execute` | Execute code |
| POST | `/api/v1/flywheel/replies/:id/verify` | Verify solution |
| POST | `/api/v1/flywheel/replies/:id/accept` | Accept as solution |
| POST | `/api/v1/flywheel/replies/:id/mark-helpful` | Upvote |

### Reputation System

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/flywheel/reputation/me` | My reputation |
| GET | `/api/v1/flywheel/reputation/:user_id` | User reputation |
| GET | `/api/v1/flywheel/reputation/:user_id/history` | History |
| GET | `/api/v1/flywheel/leaderboards/:score_type` | Leaderboards |

### Challenge System

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/flywheel/challenges` | List challenges |
| GET | `/api/v1/flywheel/challenges/:id` | Challenge details |
| POST | `/api/v1/flywheel/challenges/:id/submit` | Submit entry |
| GET | `/api/v1/flywheel/challenges/:id/leaderboard` | Leaderboard |

### Agent Collaboration

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/flywheel/threads/:id/agents/:agent_id/invite` | Invite agent |
| GET | `/api/v1/flywheel/threads/:id/agents` | List agents |
| DELETE | `/api/v1/flywheel/threads/:id/agents/:agent_id` | Remove agent |

### Executable Thread Replay

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/flywheel/threads/:id/timeline` | Get timeline |
| POST | `/api/v1/flywheel/threads/:id/replay` | Replay thread |

### Search & Discovery

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/flywheel/search` | Search |
| GET | `/api/v1/flywheel/solutions/verified` | Verified solutions |
| GET | `/api/v1/flywheel/categories` | Categories |

### Subscriptions

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/flywheel/threads/:id/subscribe` | Subscribe |
| DELETE | `/api/v1/flywheel/threads/:id/subscribe` | Unsubscribe |
| GET | `/api/v1/flywheel/subscriptions` | List subscriptions |

### Marketplace Integration

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/flywheel/replies/:id/publish-to-marketplace` | Publish function |

---

## Appendix B: Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2024-01-15 | Initial API specification |
