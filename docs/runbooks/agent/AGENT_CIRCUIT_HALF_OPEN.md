# Agent Circuit Half-Open Runbook

## Severity

Warning

## Description

Agent {{ $labels.agent_id }} circuit breaker is HALF-OPEN. The agent is in limited test traffic mode, allowing试探性请求 to check if the underlying issue is resolved.

## Understanding Half-Open State

This is a **transitional state** - the circuit breaker is testing if the agent can handle traffic again after being OPEN.

```
OPEN (1) ──► HALF-OPEN (2) ──► CLOSED (0) [success]
                           └─► OPEN (1)      [failure]
```

## Diagnosis

1. **Check circuit state history:**

   ```bash
   curl https://api.functionfly.com/v1/agents/{{ $labels.agent_id }}/circuit-history?last=1h
   ```

2. **Monitor test requests:**

   ```bash
   # Watch for test traffic
   kubectl logs -l agent_id={{ $labels.agent_id }} --since=5m | grep -i "test\|probe\|half-open"
   ```

3. **Check recent OPEN events:**

   ```bash
   ff admin agent events {{ $labels.agent_id }} --type=circuit_open --since=2h
   ```

## Common Causes

| Cause | Resolution |
|-------|------------|
| Previous failures recovered | Normal behavior; no action needed |
| Flaky downstream service | Monitor closely; may flip back to OPEN |
| Partial recovery | Investigate remaining issues |

## Remediation Steps

### Step 1: Monitor (2-5 min)

```bash
# Watch circuit state transitions
watch -n 10 'curl -s https://api.functionfly.com/v1/agents/{{ $labels.agent_id }}/circuit-state'

# Monitor success rate of test requests
ff admin agent metrics {{ $labels.agent_id }} --since=5m --format=prometheus
```

### Step 2: If Circuit Recloses (success)

No action needed - this is the desired outcome.

### Step 3: If Circuit Reopens (failure)

```bash
# Get failure reasons
ff admin agent logs {{ $labels.agent_id }} --level=error --since=5m

# Take corrective action based on root cause
# See AGENT_CIRCUIT_OPEN.md for detailed steps
```

## Prevention

- Ensure root cause of OPEN state is fully resolved before half-open test
- Implement gradual traffic increase after recovery
- Monitor closely during half-open phase
