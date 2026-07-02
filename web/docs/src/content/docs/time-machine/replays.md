---
title: Replays
description: Scan execution history, replay through corrected code, and diff outputs
sidebar:
  order: 2
---


A **replay** scans a function's execution history within a time window, re-runs
every request through a corrected version, and diffs the old vs. new outputs.

## Creating a Replay

### Dashboard

1. Navigate to **Time Machine → New Replay**
2. Select the function to replay
3. Set the **time window** (start and end)
4. Choose the **target version** (the corrected function version)
5. Set **max executions** (optional cap)
6. Choose **reconciliation mode**: `dry_run` (preview only), `live`, or `preview_only`
7. Click **Start Replay**

### API

```bash
curl -X POST https://api.functionfly.com/v1/time-machine/replays \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "fx_abc123",
    "window_start": "2026-06-20T00:00:00Z",
    "window_end": "2026-06-23T00:00:00Z",
    "target_version_id": "v2.0.0",
    "max_executions": 5000,
    "reconciliation_mode": "dry_run",
    "reason": "Fix incorrect currency conversion rate",
    "incident_url": "https://status.example.com/incidents/123"
  }'
```

### Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `function_id` | string | Yes | The function to replay |
| `window_start` | ISO 8601 | Yes | Start of the time window |
| `window_end` | ISO 8601 | Yes | End of the time window |
| `target_version_id` | string | Yes | The corrected function version to replay through |
| `max_executions` | int | No | Cap on executions to replay (default: plan limit) |
| `reconciliation_mode` | string | No | `dry_run` (default), `live`, or `preview_only` |
| `reason` | string | No | Human-readable reason for the replay |
| `incident_url` | string | No | Link to the incident ticket |

## Pipeline Phases

A replay progresses through these phases:

### 1. Scanning

The engine queries the execution history for the function within the specified
window. It identifies every execution that ran against the old version.

Progress is published in real-time:

```
[scanning] Found 2,847 executions in window (68% scanned)
```

### 2. Replaying

Each original request is fed through the corrected function version in an
isolated sandbox. The new output is captured alongside the original.

```
[replaying] 1,423/2,847 executions replayed (50%)
```

### 3. Diffing

Every pair (original output, new output) is compared field-by-field:

| Classification | Meaning |
|---|---|
| `identical` | Outputs match exactly |
| `minor` | Cosmetic differences (whitespace, formatting) |
| `major` | Semantic differences in values |
| `breaking` | Schema change or missing fields |
| `error` | New version returned an error for this input |

```
[diffing] 2,847 items: 2,100 identical, 412 minor, 287 major, 48 breaking
```

### 4. Completed

The replay is finished. You can now:
- Review individual items and their diffs
- View the aggregate diff summary
- Start reconciliation (Pro+)
- Generate an audit certificate (Enterprise)

## Monitoring Progress

### Dashboard

The replay detail page shows a real-time progress bar and phase indicator.
Active replays auto-refresh every 5 seconds.

### SSE Stream

```bash
curl -N https://api.functionfly.com/v1/time-machine/replays/{id}/stream \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

Events:

```
data: {"phase":"scanning","progress":68,"found":2847}
data: {"phase":"replaying","progress":50,"replayed":1423,"total":2847}
data: {"phase":"diffing","progress":100,"identical":2100,"minor":412,"major":287,"breaking":48}
data: {"phase":"completed","progress":100}
```

### Polling

```bash
curl https://api.functionfly.com/v1/time-machine/replays/{id}/progress \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
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

## Listing Replays

```bash
curl https://api.functionfly.com/v1/time-machine/replays \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

Returns all replay jobs for the current tenant, sorted by creation date.

## Cancelling a Replay

An in-progress replay can be cancelled at any time:

```bash
curl -X DELETE https://api.functionfly.com/v1/time-machine/replays/{id} \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

Cancelled replays retain all items processed up to the cancellation point.

## Diff Summary

After a replay completes, get the aggregate diff breakdown:

```bash
curl https://api.functionfly.com/v1/time-machine/replays/{id}/diff-summary \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
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

## Reviewing Individual Items

List all replay items with their diff classification:

```bash
curl https://api.functionfly.com/v1/time-machine/replays/{id}/items \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

Get a single item with full diff detail:

```bash
curl https://api.functionfly.com/v1/time-machine/replays/{id}/items/{itemId} \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

```json
{
  "id": "item_001",
  "original_execution_id": "exec_xyz",
  "original_input": { "amount": 100, "currency": "EUR" },
  "original_output": { "converted": 108.50 },
  "new_output": { "converted": 112.30 },
  "diff_type": "major",
  "diff_summary": "converted: 108.50 → 112.30",
  "diff_detail": {
    "fields_changed": [
      { "path": "converted", "old": 108.50, "new": 112.30, "classification": "major" }
    ]
  }
}
```

## Next Steps

- [Reconciliation](/time-machine/reconciliation/) — Apply corrections to real-world state
- [Audit Certificates](/time-machine/audit-certificates/) — Generate compliance proofs
- [API Reference](/time-machine/api/) — Full endpoint docs
