# Agent Quota Violations Runbook

## Severity

Warning (> 5/sec)

## Description

Agent {{ $labels.agent_id }} is generating quota violations at an elevated rate, indicating the agent is consistently exceeding its allocated limits.

## Diagnosis

1. **Check quota violation details:**

   ```bash
   ff admin agent quota-violations {{ $labels.agent_id }} --since=1h
   ```

2. **Identify which quota is being exceeded:**

   ```bash
   ff admin agent quota {{ $labels.agent_id }} --status --format=table
   ```

3. **Check violation patterns:**

   ```bash
   ff admin agent quota-violations {{ $labels.agent_id }} --by-minute --since=1h
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| Quota limits too low | Common | Increase quota or optimize usage |
| Agent not respecting limits | Rare | Check quota enforcement code |
| Traffic spike | Common | Rate limit incoming traffic |
| Misconfigured quota | Common | Review and fix quota config |

## Remediation Steps

### Step 1: Immediate Action (2 min)

```bash
# View current quota config
ff admin agent quota {{ $labels.agent_id }} --config

# Temporarily increase quota
ff admin agent quota {{ $labels.agent_id }} --increase=50% --duration=4h
```

### Step 2: Investigate (10 min)

```bash
# Get violation breakdown
ff admin agent quota-violations {{ $labels.agent_id }} --format=json > violations.json

# Check if this is consistent or spike
cat violations.json | jq '.[] | .quota_type' | sort | uniq -c
```

### Step 3: Long-term Fix

- If **limits too low**: Adjust quota to match actual usage
- If **traffic spike**: Implement rate limiting at API gateway
- If **agent bug**: File bug; temporarily increase quota

## Prevention

- Set up proactive monitoring at 80% quota usage
- Review quota usage weekly
- Implement automatic quota adjustment based on usage patterns
- Use per-endpoint quotas to prevent single endpoint exhaustion
