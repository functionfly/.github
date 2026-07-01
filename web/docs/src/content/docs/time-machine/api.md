---
title: Time Machine API
description: Full API reference for the FunctionFly Time Machine
sidebar:
  order: 5
---

# Time Machine API Reference

All endpoints are under `/v1/time-machine` and require authentication.
Feature gating is applied per-plan.

## Endpoints

### Replays

| Method | Path | Description | Min Plan |
|---|---|---|---|
| `POST` | `/replays` | Create a replay job | Free |
| `GET` | `/replays` | List all replay jobs | Free |
| `GET` | `/replays/{id}` | Get a single replay | Free |
| `DELETE` | `/replays/{id}` | Cancel a replay | Free |
| `GET` | `/replays/{id}/progress` | Get replay progress | Free |
| `GET` | `/replays/{id}/stream` | SSE progress stream | Free |
| `GET` | `/replays/{id}/items` | List replay items | Free |
| `GET` | `/replays/{id}/items/{itemId}` | Get a single item | Free |
| `GET` | `/replays/{id}/diff-summary` | Aggregate diff breakdown | Free |

### Reconciliation

| Method | Path | Description | Min Plan |
|---|---|---|---|
| `POST` | `/replays/{id}/reconcile` | Start reconciliation | Pro |
| `GET` | `/replays/{id}/reconciliations` | List reconciliations | Pro |

### Audit

| Method | Path | Description | Min Plan |
|---|---|---|---|
| `GET` | `/replays/{id}/audit-certificate` | Get audit certificate | Enterprise |

### Limits

| Method | Path | Description | Min Plan |
|---|---|---|---|
| `GET` | `/limits` | Get plan limits for Time Machine | Free |

---

## Create Replay

```
POST /v1/time-machine/replays
```

### Request Body

```json
{
  "function_id": "fx_abc123",
  "window_start": "2026-06-20T00:00:00Z",
  "window_end": "2026-06-23T00:00:00Z",
  "target_version_id": "v2.0.0",
  "max_executions": 5000,
  "reconciliation_mode": "dry_run",
  "reason": "Fix incorrect currency conversion rate",
  "incident_url": "https://status.example.com/incidents/123"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `function_id` | string | Yes | Function to replay |
| `window_start` | ISO 8601 | Yes | Start of time window |
| `window_end` | ISO 8601 | Yes | End of time window |
| `target_version_id` | string | Yes | Corrected version to replay through |
| `max_executions` | int | No | Cap on executions (default: plan limit) |
| `reconciliation_mode` | string | No | `dry_run` (default), `live`, `preview_only` |
| `reason` | string | No | Human-readable reason |
| `incident_url` | string | No | Link to incident ticket |

### Response

```json
{
  "id": "replay_abc123",
  "tenant_id": "ten_xyz",
  "function_id": "fx_abc123",
  "status": "pending",
  "window_start": "2026-06-20T00:00:00Z",
  "window_end": "2026-06-23T00:00:00Z",
  "target_version_id": "v2.0.0",
  "reconciliation_mode": "dry_run",
  "created_at": "2026-06-23T14:00:00Z"
}
```

### Errors

| Status | Code | Meaning |
|---|---|---|
| 400 | `INVALID_WINDOW` | Window exceeds plan limit |
| 400 | `INVALID_VERSION` | Target version not found |
| 403 | `FEATURE_NOT_AVAILABLE` | Plan doesn't include this feature |
| 409 | `MAX_CONCURRENT` | Too many active replays |
| 429 | `RATE_LIMITED` | Rate limit exceeded |

---

## List Replays

```
GET /v1/time-machine/replays
```

### Query Parameters

| Param | Type | Default | Description |
|---|---|---|---|
| `function_id` | string | — | Filter by function |
| `status` | string | — | Filter by status |
| `limit` | int | 50 | Page size (max 100) |
| `offset` | int | 0 | Pagination offset |

### Response

```json
{
  "replays": [
    {
      "id": "replay_abc123",
      "function_id": "fx_abc123",
      "status": "completed",
      "progress_percent": 100,
      "total_executions_found": 2847,
      "total_executions_replayed": 2847,
      "total_executions_changed": 747,
      "created_at": "2026-06-23T14:00:00Z",
      "completed_at": "2026-06-23T14:12:34Z"
    }
  ],
  "total": 1
}
```

---

## Get Replay

```
GET /v1/time-machine/replays/{id}
```

Returns the full replay object including progress counters.

---

## Cancel Replay

```
DELETE /v1/time-machine/replays/{id}
```

Cancels an in-progress replay. Items processed so far are retained.

---

## Replay Progress

```
GET /v1/time-machine/replays/{id}/progress
```

```json
{
  "status": "replaying",
  "progress_percent": 50,
  "total_executions_found": 2847,
  "total_executions_replayed": 1423,
  "total_executions_changed": 522,
  "total_executions_failed": 3
}
```

---

## SSE Stream

```
GET /v1/time-machine/replays/{id}/stream
```

Returns `text/event-stream`. Events:

```
data: {"phase":"scanning","progress":100,"found":2847}
data: {"phase":"replaying","progress":50,"replayed":1423,"total":2847}
data: {"phase":"diffing","progress":100,"identical":2100,"minor":412,"major":287,"breaking":48}
data: {"phase":"completed","progress":100}
```

---

## List Replay Items

```
GET /v1/time-machine/replays/{id}/items
```

### Query Parameters

| Param | Type | Default | Description |
|---|---|---|---|
| `diff_type` | string | — | Filter: `identical`, `minor`, `major`, `breaking`, `error` |
| `limit` | int | 50 | Page size (max 200) |
| `offset` | int | 0 | Pagination offset |

---

## Get Replay Item

```
GET /v1/time-machine/replays/{id}/items/{itemId}
```

Returns the full item with `original_input`, `original_output`, `new_output`,
`diff_type`, `diff_summary`, and `diff_detail`.

---

## Diff Summary

```
GET /v1/time-machine/replays/{id}/diff-summary
```

```json
{
  "total_items": 2847,
  "identical": 2100,
  "minor": 412,
  "major": 287,
  "breaking": 48,
  "error": 0,
  "percent_changed": 26.3
}
```

---

## Start Reconciliation

```
POST /v1/time-machine/replays/{id}/reconcile
```

Requires `time_machine_pro` feature or above.

### Request Body

```json
{
  "mode": "dry_run",
  "include_minor": false,
  "include_major": true,
  "include_breaking": true
}
```

### Response

```json
{
  "reconciliation_id": "rec_abc",
  "mode": "dry_run",
  "total_actions": 335,
  "actions": [...],
  "estimated_impact": {
    "executions_modified": 335,
    "downstream_notifications": 12
  }
}
```

---

## List Reconciliations

```
GET /v1/time-machine/replays/{id}/reconciliations
```

---

## Audit Certificate

```
GET /v1/time-machine/replays/{id}/audit-certificate
```

Requires `time_machine_enterprise` feature.

Returns the full audit certificate with Merkle tree, Ed25519 signature,
compliance framework tags, and retention policy.

---

## Get Limits

```
GET /v1/time-machine/limits
```

Returns the current plan's Time Machine limits:

```json
{
  "replay_window_hours": 720,
  "max_executions_per_replay": 10000,
  "max_concurrent_replays": 3,
  "dry_run_reconciliation": true,
  "auto_reconciliation": false,
  "live_reconciliation": true,
  "audit_certificates": false,
  "replay_scheduling": false
}
```

---

## Replay Status Values

| Status | Description |
|---|---|
| `pending` | Queued, waiting to start |
| `scanning` | Finding executions in the time window |
| `replaying` | Running executions through the new version |
| `diffing` | Comparing old vs. new outputs |
| `reconciling` | Applying reconciliation actions |
| `completed` | Finished successfully |
| `failed` | Encountered an unrecoverable error |
| `cancelled` | Cancelled by the user |

## Error Codes

| Status | Code | Meaning |
|---|---|---|
| 400 | `INVALID_WINDOW` | Time window exceeds plan limit |
| 400 | `INVALID_VERSION` | Target version not found |
| 400 | `INVALID_MODE` | Invalid reconciliation mode |
| 403 | `FEATURE_NOT_AVAILABLE` | Plan doesn't include this feature |
| 404 | `REPLAY_NOT_FOUND` | Replay ID not found |
| 404 | `ITEM_NOT_FOUND` | Item ID not found |
| 409 | `MAX_CONCURRENT` | Too many concurrent replays |
| 409 | `ALREADY_COMPLETED` | Replay already finished |
| 429 | `RATE_LIMITED` | Rate limit exceeded |
