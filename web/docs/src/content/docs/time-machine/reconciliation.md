---
title: Reconciliation
description: Apply corrections to real-world state after a replay
sidebar:
  order: 3
---

# Reconciliation

After a replay completes and you've reviewed the diffs, **reconciliation**
applies corrections to the real-world side effects of the original executions.

## Modes

| Mode | Description | Available |
|---|---|---|
| `dry_run` | Preview what would change without modifying anything | All plans |
| `preview_only` | Generate a detailed plan but do not apply | All plans |
| `live` | Apply corrections to real state | Pro+ |

Always start with `dry_run` to review the reconciliation plan before going live.

## Starting Reconciliation

### Dashboard

1. Open a completed replay
2. Review the diff summary
3. Click **Reconcile**
4. Choose mode (`dry_run` or `live`)
5. Review the action list
6. Confirm to apply (live mode)

### API

```bash
curl -X POST https://api.functionfly.com/v1/time-machine/replays/{id}/reconcile \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "dry_run",
    "include_minor": false,
    "include_major": true,
    "include_breaking": true
  }'
```

### Request Fields

| Field | Type | Description |
|---|---|---|
| `mode` | string | `dry_run`, `preview_only`, or `live` |
| `include_minor` | bool | Include minor diffs in reconciliation (default: false) |
| `include_major` | bool | Include major diffs (default: true) |
| `include_breaking` | bool | Include breaking diffs (default: true) |

## Reconciliation Actions

The engine generates one action per changed execution:

| Action Type | Description |
|---|---|
| `update_output` | Replace the stored output with the new (corrected) output |
| `flag_error` | Mark the original execution as errored in the audit log |
| `notify_downstream` | Send a webhook notification to downstream consumers |
| `append_correction` | Append a correction record without overwriting the original |

Each action includes:

```json
{
  "action_type": "update_output",
  "target_resource": "exec_xyz",
  "old_value": { "converted": 108.50 },
  "new_value": { "converted": 112.30 },
  "dry_run": true,
  "reversible": true,
  "reversal_data": { "original_output": { "converted": 108.50 } }
}
```

## Dry Run Results

A dry run returns the full plan without applying changes:

```json
{
  "reconciliation_id": "rec_abc",
  "mode": "dry_run",
  "total_actions": 335,
  "actions": [
    {
      "action_type": "update_output",
      "target_resource": "exec_xyz",
      "old_value": { "converted": 108.50 },
      "new_value": { "converted": 112.30 },
      "dry_run": true
    }
  ],
  "estimated_impact": {
    "executions_modified": 335,
    "downstream_notifications": 12
  }
}
```

Review the plan carefully. When satisfied, start a live reconciliation with
the same parameters (change `mode` to `live`).

## Live Reconciliation

Live mode applies every action in the plan. The process:

1. Actions are applied in order (sorted by original execution timestamp)
2. Each action is atomic — if one fails, the rest continue
3. Reversal data is stored for every action (reversible by default)
4. Progress is published via SSE, identical to the replay stream

:::caution[Live reconciliation modifies real state]
Once applied, the original execution outputs in the registry are updated.
Reversal data is stored, but you should always review the dry run first.
:::

## Reversibility

All reconciliation actions store reversal data. If you need to undo a
reconciliation, contact support with the reconciliation ID. The reversal
replays the original outputs back into place.

## Listing Reconciliations

```bash
curl https://api.functionfly.com/v1/time-machine/replays/{id}/reconciliations \
  -H "Authorization: Bearer $FUNCTIONFLY_API_KEY"
```

## Billing

Live reconciliation triggers a metered usage event (`time_machine_replay`)
based on the number of executions reconciled. Dry runs are free.

## Next Steps

- [Audit Certificates](/time-machine/audit-certificates/) — Generate compliance proofs for reconciliations
- [API Reference](/time-machine/api/) — Full endpoint docs
