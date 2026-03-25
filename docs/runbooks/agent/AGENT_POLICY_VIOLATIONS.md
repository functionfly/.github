# Agent Policy Violations Runbook

## Severity

Warning (> 10/sec)

## Description

Agent {{ $labels.agent_id }} is generating policy violations at an elevated rate.

## Diagnosis

1. **Check violation types:**

   ```bash
   ff admin agent violations {{ $labels.agent_id }} --since=1h --group-by=type
   ```

2. **Examine recent violations:**

   ```bash
   ff admin agent violations {{ $labels.agent_id }} --recent --limit=50
   ```

3. **Check security service logs:**

   ```bash
   kubectl logs -l agent_id={{ $labels.agent_id }} --container=security --since=30m | grep -i violation
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| Prompt injection attack | Rare | Investigate source; block attacker |
| Agent malfunction | Rare | Rollback agent version |
| Malicious input | Common | Add input sanitization |
| Misconfigured policy | Common | Adjust policy thresholds |

## Remediation Steps

### Step 1: Stop the Bleeding (2 min)

```bash
# Temporarily suspend agent
ff admin agent suspend {{ $labels.agent_id }} --reason="policy-violations"

# Check if this is an attack or malfunction
ff admin agent violations {{ $labels.agent_id }} --source-analysis
```

### Step 2: Analyze Violations (10 min)

```bash
# Get violation details
ff admin agent violations {{ $labels.agent_id }} --format=json > violations.json

# Categorize
cat violations.json | jq -r '.[].policy_type' | sort | uniq -c | sort -rn
```

### Step 3: Fix and Resume

- If **attack**: Block source IP, add WAF rule
- If **malfunction**: Rollback agent, file bug
- If **input issue**: Add validation
- If **policy too strict**: Adjust thresholds

### Step 4: Resume Agent

```bash
ff admin agent resume {{ $labels.agent_id }}
```

## Prevention

- Implement input sanitization
- Use WAF to block prompt injection patterns
- Regular policy review
- Anomaly detection on violation patterns
