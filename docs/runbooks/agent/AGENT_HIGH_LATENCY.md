# Agent High Latency Runbook

## Severity

Warning (P95 > 5s), Critical (P99 > 10s)

## Description

Agent {{ $labels.agent_id }} is experiencing high latency. P95 latency exceeds 5 seconds.

## Diagnosis

1. **Check current latency:**

   ```bash
   curl https://api.functionfly.com/v1/agents/{{ $labels.agent_id }}/metrics --metric=latency
   ```

2. **Identify slow operations:**

   ```bash
   ff admin agent traces {{ $labels.agent_id }} --slowest --limit=20
   ```

3. **Check resource utilization:**

   ```bash
   # CPU/Memory
   kubectl top pod -l agent_id={{ $labels.agent_id }}
   
   # Network
   kubectl exec -it agent/{{ $labels.agent_id }} -- cat /proc/net/dev
   ```

4. **Check LLM provider latency:**

   ```bash
   curl https://api.openrouter.ai/v1/models | jq '.data[] | select(.id=="anthropic/claude-3-sonnet") | {id, latency:p99_latency_ms}'
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| LLM provider slow | Common | Check provider status; try alternative |
| Network latency | Common | Check network route; use CDN |
| Cold starts | Common | Keep agent warm; reduce package size |
| Database queries slow | Common | Add indexes; optimize queries |
| Memory pressure | Common | Increase memory limit; reduce concurrency |

## Remediation Steps

### Step 1: Identify Bottleneck (5 min)

```bash
# Get distributed traces
ff admin agent traces {{ $labels.agent_id }} --since=5m --format=json > traces.json

# Analyze latency breakdown
cat traces.json | jq -r '.spans[] | .name' | sort | uniq -c | sort -rn | head
```

### Step 2: Quick Fixes

```bash
# Warm up agent (reduce cold starts)
ff admin agent warm {{ $labels.agent_id }}

# Reduce concurrency to relieve pressure
ff admin agent scale {{ $labels.agent_id }} --concurrency=5

# Enable caching if available
ff admin agent config {{ $labels.agent_id }} --set=caching.enabled=true
```

### Step 3: Long-term Fixes

- Optimize function package size
- Add caching layer
- Scale horizontally
- Use faster LLM model

## Prevention

- Set SLO: P95 < 2s, P99 < 5s
- Monitor latency trends daily
- Use latency budgets
- Implement synthetic monitoring
