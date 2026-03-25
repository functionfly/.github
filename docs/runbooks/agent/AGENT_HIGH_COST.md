# Agent High Cost Runbook

## Severity

Warning (>$10/hour)

## Description

Agent {{ $labels.agent_id }} is incurring costs at a rate exceeding $10/hour, which may indicate runaway usage or budget misconfiguration.

## Diagnosis

1. **Check current cost rate:**

   ```bash
   ff admin agent cost {{ $labels.agent_id }} --rate --current
   ```

2. **View cost breakdown:**

   ```bash
   ff admin agent cost {{ $labels.agent_id }} --breakdown --since=24h
   ```

3. **Identify cost drivers:**

   ```bash
   ff admin agent cost {{ $labels.agent_id }} --by-function --since=24h | sort -rn | head
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| LLM model too expensive | Common | Switch to cheaper model |
| Excessive function calls | Common | Optimize; cache responses |
| runaway loop | Rare | Suspend agent; investigate |
| Misconfigured budget | Common | Adjust budget limits |
| Unexpected traffic spike | Common | Rate limit; analyze traffic |

## Remediation Steps

### Step 1: Stop the Bleeding (2 min)

```bash
# Set emergency cost cap
ff admin agent cost {{ $labels.agent_id }} --cap=50 --duration=24h

# OR temporarily suspend agent
ff admin agent suspend {{ $labels.agent_id }} --reason="high-cost"
```

### Step 2: Analyze Costs (10 min)

```bash
# Get detailed cost breakdown
ff admin agent cost {{ $labels.agent_id }} --format=json > costs.json

# Analyze by component
cat costs.json | jq '.cost_breakdown | to_entries | sort_by(.value) | reverse'
```

### Step 3: Implement Fixes

```bash
# Switch to cheaper LLM model
ff admin agent config {{ $labels.agent_id }} --set=llm.model=claude-3-haiku

# Enable response caching
ff admin agent config {{ $labels.agent_id }} --set=caching.enabled=true

# Set spending alert
ff admin agent cost {{ $labels.agent_id }} --alert=80%
```

### Step 4: Set Budget Controls

```bash
# Set daily budget
ff admin agent budget {{ $labels.agent_id }} --daily=100

# Enable auto-suspend on budget
ff admin agent budget {{ $labels.agent_id }} --auto-suspend=true
```

## Prevention

- Set up cost budgets with alerts
- Use cost-effective LLM models by default
- Implement caching aggressively
- Review cost trends weekly
- Use per-function cost tracking
