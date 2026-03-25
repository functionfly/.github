# Agent Quota Exhausted Runbook

## Severity

Warning

## Description

Agent {{ $labels.agent_id }} has used over 90% of its allocated quota.

## Diagnosis

1. **Check quota usage:**

   ```bash
   ff admin agent quota {{ $labels.agent_id }} --current
   ```

2. **View quota consumption breakdown:**

   ```bash
   ff admin agent quota {{ $labels.agent_id }} --breakdown --since=24h
   ```

3. **Identify what's consuming quota:**

   ```bash
   ff admin agent usage {{ $labels.agent_id }} --format=json | jq '.items | sort_by(.cost) | reverse | .[:10]'
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| Unexpected traffic spike | Common | Review traffic patterns; adjust quota |
| runaway loop in agent | Rare | Suspend agent; investigate |
| Quota limit too low | Common | Increase quota or optimize usage |
| LLM provider price increase | Rare | Review costs; switch provider |

## Remediation Steps

### Step 1: Assess Urgency (2 min)

```bash
# Check if quota is still increasing
ff admin agent quota {{ $labels.agent_id }} --watch --interval=30s

# Estimate time until exhaustion
ff admin agent quota {{ $labels.agent_id }} --time-to-exhaust
```

### Step 2: Emergency Quota Increase (if critical)

```bash
# Temporarily increase quota to prevent service disruption
ff admin agent quota {{ $labels.agent_id }} --increase=2x --duration=24h

# OR disable quota limits temporarily
ff admin agent quota {{ $labels.agent_id }} --suspend --reason="emergency"
```

### Step 3: Investigate Root Cause (15 min)

```bash
# Get top consumers
ff admin agent usage {{ $labels.agent_id }} --since=1h --group-by=function

# Check for anomalous patterns
ff admin agent usage {{ $labels.agent_id }} --since=1h --compare=daily
```

### Step 4: Long-term Fix

```bash
# Adjust quota limits
ff admin agent quota {{ $labels.agent_id }} --set-daily=10000 --set-monthly=100000

# OR optimize agent to use fewer resources
ff admin agent optimize {{ $labels.agent_id }}
```

## Prevention

- Set up proactive alerts at 75% quota usage
- Review quota usage weekly
- Implement automatic scaling for quota
- Use cost allocation tags to track usage by team
