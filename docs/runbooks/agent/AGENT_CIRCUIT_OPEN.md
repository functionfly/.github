# Agent Circuit Open Runbook

## Severity

Critical

## Description

Agent {{ $labels.agent_id }} circuit breaker is OPEN. The agent is temporarily unavailable due to repeated failures.

## Understanding Circuit Breaker States

```
CLOSED (0) ──► OPEN (1) ──► HALF-OPEN (2) ──► CLOSED (0)
    │              │              │
    │ Normal       │ Too many     │ Test request
    │ operation    │ failures     │ succeeds
    │              │              │
    └──────────────┴──────────────┘
```

## Diagnosis

1. **Check circuit breaker state:**

   ```bash
   curl https://api.functionfly.com/v1/agents/{{ $labels.agent_id }}/circuit-state
   ```

2. **Identify failure causes:**

   ```bash
   # Check recent errors
   kubectl logs -l agent_id={{ $labels.agent_id }} --since=1h | grep -i error | tail -50
   
   # Check if downstream service is down
   curl -I https://api.openrouter.ai/health
   ```

3. **Check circuit breaker configuration:**

   ```bash
   ff admin agent config {{ $labels.agent_id }} --get=circuit_breaker
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| Downstream LLM provider down | Rare | Wait for provider recovery; switch backup |
| Agent overwhelmed by traffic | Common | Scale agent; implement queue |
| Code bug causing 100% failure | Common | Rollback; fix bug |
| Network partition | Rare | Wait for network recovery |
| Dependency timeout | Common | Increase timeout; fix slow dependency |

## Remediation Steps

### Step 1: Assess Situation (2 min)

```bash
# Check if this is part of a larger outage
ff admin status

# Check provider status
curl https://status.openrouter.ai/api/v2/status.svg
```

### Step 2: Wait for Auto-Recovery (5 min)

Circuit breakers automatically transition to HALF-OPEN after the recovery timeout. Default is 60 seconds.

```bash
# Watch circuit state
watch -n 5 'curl -s https://api.functionfly.com/v1/agents/{{ $labels.agent_id }}/circuit-state'
```

### Step 3: Manual Reset (if urgent)

```bash
# Force circuit breaker reset (agent will immediately receive traffic)
ff admin agent circuit-reset {{ $labels.agent_id }}

# OR gracefully disable agent first
ff admin agent disable {{ $labels.agent_id }} --reason="circuit-open-manual-reset"
ff admin agent circuit-reset {{ $labels.agent_id }}
ff admin agent enable {{ $labels.agent_id }}
```

### Step 4: Investigate Root Cause

```bash
# Get failure logs
ff admin agent logs {{ $labels.agent_id }} --level=error --since=1h > circuit_errors.log

# Analyze failure patterns
cat circuit_errors.log | jq -r '.error.code' | sort | uniq -c | sort -rn
```

## Tuning Circuit Breaker

If circuit breaker is triggering too aggressively:

```bash
# View current config
ff admin agent config {{ $labels.agent_id }} --get=circuit_breaker

# Adjust failure threshold (default: 5 failures)
ff admin agent config {{ $labels.agent_id }} --set=circuit_breaker.failure_threshold=10

# Adjust recovery timeout (default: 60s)
ff admin agent config {{ $labels.agent_id }} --set=circuit_breaker.recovery_timeout=120s
```

## Prevention

- Monitor downstream service health
- Set up redundancy with multiple LLM providers
- Implement proper timeout and retry budgets
- Use chaos engineering to test circuit breaker behavior
