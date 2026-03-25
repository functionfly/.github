# Agent No Executions Runbook

## Severity

Info

## Description

Agent {{ $labels.agent_id }} has not executed any functions in the last 15 minutes. This may indicate the agent is idle, misconfigured, or blocked.

## Diagnosis

1. **Check agent status:**

   ```bash
   ff admin agent status {{ $labels.agent_id }}
   ```

2. **Verify agent is enabled:**

   ```bash
   ff admin agent get {{ $labels.agent_id }} --field=status
   ```

3. **Check recent events:**

   ```bash
   ff admin agent events {{ $labels.agent_id }} --since=30m
   ```

4. **Test agent manually:**

   ```bash
   ff admin agent invoke {{ $labels.agent_id }} --input='{"test": true}'
   ```

## Common Causes

| Cause | Resolution |
|-------|------------|
| Agent disabled/suspended | Enable agent |
| No incoming requests | Check traffic; verify routing |
| All requests failing silently | Review recent errors |
| Agent in cooldown | Wait for cooldown to expire |
| Misconfigured trigger | Check trigger configuration |

## Remediation Steps

### Step 1: Check Agent Status (2 min)

```bash
# Get full agent status
ff admin agent status {{ $labels.agent_id }} --verbose

# Check if agent is accepting work
ff admin agent get {{ $labels.agent_id }} --field=accepting_tasks
```

### Step 2: Test Agent (2 min)

```bash
# Invoke test
ff admin agent invoke {{ $labels.agent_id }} --input='{"test": true}' --wait

# Check execution history
ff admin agent executions {{ $labels.agent_id }} --last=5
```

### Step 3: Fix Issues

- If **disabled**: `ff admin agent enable {{ $labels.agent_id }}`
- If **no requests**: Check API gateway logs; verify routing
- If **test fails**: Investigate failure reason
- If **cooldown**: Wait or `ff admin agent cooldown-reset {{ $labels.agent_id }}`

### Step 4: Monitor

```bash
# Watch for new executions
ff admin agent watch {{ $labels.agent_id }} --timeout=60s
```

## Prevention

- Set up uptime monitoring
- Implement health check endpoints
- Use synthetic transactions to verify agent is working
- Alert on extended idle periods
