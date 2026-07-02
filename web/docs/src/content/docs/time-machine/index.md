---
title: Time Machine
description: Retroactively fix bugs in production by replaying real executions through corrected code
---


**Time Machine** lets you retroactively fix production bugs by replaying real
executions through a corrected function version, diffing old vs. new outputs,
and reconciling the real-world side effects — all with a full audit trail.

This is a FunctionFly-exclusive feature. No other serverless platform offers
retroactive execution correction with compliance-grade audit certificates.

## How It Works

```
  SCAN           REPLAY           DIFF           RECONCILE        AUDIT
┌────────┐    ┌──────────┐    ┌──────────┐    ┌────────────┐   ┌──────────┐
│ Find   │    │ Run old  │    │ Compare  │    │ Apply      │   │ Merkle   │
│ execs  │───►│ requests │───►│ outputs  │───►│ fixes to   │──►│ tree     │
│ in     │    │ through  │    │ field by │    │ live state │   │ + Ed25519│
│ window │    │ new code │    │ field    │    │            │   │ cert     │
└────────┘    └──────────┘    └──────────┘    └────────────┘   └──────────┘
```

1. **Scan** — Find every execution of a function within a time window
2. **Replay** — Run each original request through your corrected function version
3. **Diff** — Compare old vs. new outputs (classified as identical, minor, major, breaking, or error)
4. **Reconcile** — Apply corrections to real-world state (dry-run or live)
5. **Audit** — Generate a signed, tamper-proof audit certificate

## When to Use Time Machine

- **Production bug fix** — A function returned wrong data for 3 days. Fix the code, replay all affected requests, and correct the outputs.
- **Data pipeline correction** — A transformation function had a subtle error that corrupted downstream data. Replay and reconcile.
- **Compliance incident** — An auditor asks "what was affected by bug X?" Generate a full audit certificate with every affected execution.

## Plan Tiers

| Capability | Free | Starter | Pro | Enterprise |
|---|---|---|---|---|
| Replay window | 24 hours | 72 hours | 30 days | 90 days |
| Max executions per replay | 100 | 1,000 | 10,000 | 100,000 |
| Concurrent replays | 1 | 1 | 3 | 10 |
| Diff analysis | Basic | Full | Full | Full |
| Dry-run reconciliation | Yes | Yes | Yes | Yes |
| Live reconciliation | — | — | Yes | Yes |
| Audit certificates | — | — | — | Yes |
| Scheduled replays | — | — | — | Yes |

## Quick Start

### From the Dashboard

1. Go to **Time Machine** in the sidebar
2. Click **New Replay**
3. Select a function and time window
4. Choose the corrected version
5. Run the replay and review diffs

### From the API

```bash
curl -X POST https://api.functionfly.com/v1/time-machine/replays \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "fx_abc123",
    "window_start": "2026-06-20T00:00:00Z",
    "window_end": "2026-06-23T00:00:00Z",
    "target_version_id": "v2.0.0",
    "reconciliation_mode": "dry_run"
  }'

curl -N https://api.functionfly.com/v1/time-machine/replays/{id}/stream \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

## Next Steps

- [Replays](/time-machine/replays/) — Scanning, replaying, and diffing in detail
- [Reconciliation](/time-machine/reconciliation/) — Fixing side effects (dry-run and live)
- [Audit Certificates](/time-machine/audit-certificates/) — Compliance-grade proof
- [API Reference](/time-machine/api/) — Full endpoint documentation
