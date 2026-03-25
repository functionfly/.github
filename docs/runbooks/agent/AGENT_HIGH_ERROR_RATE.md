# Agent High Error Rate Runbook

## Severity

Warning

## Description

Agent {{ $labels.agent_id }} has an error rate exceeding 10% over a 5-minute window.

## Diagnosis

1. **Check recent logs:**

   ```bash
   kubectl logs -l agent_id={{ $labels.agent_id }} --tail=100 | grep -i error
   ```

2. **Identify error patterns:**

   ```bash
   kubectl logs -l agent_id={{ $labels.agent_id }} --since=15m | grep -E "(ERROR|error|Error)" | cut -d: -f4 | sort | uniq -c | sort -rn | head
   ```

3. **Check agent health endpoint:**

   ```bash
   curl https://api.functionfly.com/v1/agents/{{ $labels.agent_id }}/health
   ```

4. **Examine recent deployments:**

   ```bash
   ff admin agent history {{ $labels.agent_id }} --last=10
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| LLM provider outage | Rare | Check OpenRouter status; wait for auto-recovery |
| Invalid function input | Common | Review input validation; update function schema |
| Dependency timeout | Common | Increase timeout settings; check network |
| Rate limiting | Common | Implement backoff; check quota limits |
| Code bug introduced | Common | Rollback to previous version |

## Remediation Steps

### Step 1: Isolate the Problem (2 min)

```bash
# Disable the agent to prevent cascade failures
ff admin agent disable {{ $labels.agent_id }} --reason="high-error-rate-runbook"
```

### Step 2: Analyze Errors (5 min)

```bash
# Get error distribution
ff admin agent errors {{ $labels.agent_id }} --since=1h --format=json > errors.json

# Identify error types
cat errors.json | jq -r '.errors[] | .type' | sort | uniq -c
```

### Step 3: Apply Fix

- If **LLM provider**: Wait or switch to backup provider
- If **invalid input**: Update input validation, redeploy
- If **timeout**: Increase timeout, check dependencies
- If **rate limit**: Implement exponential backoff
- If **code bug**: `ff admin agent rollback {{ $labels.agent_id }}`

### Step 4: Re-enable

```bash
ff admin agent enable {{ $labels.agent_id }}
```

## Escalation

| Time | Action |
|------|--------|
| 15 min | Page on-call if error rate > 50% |
| 30 min | Escalate to platform team lead |
| 60 min | Initiate incident response |

## Prevention

- Set up SLO: 99.5% success rate
- Monitor error rate in Grafana dashboard
- Set up pre-production validation tests
