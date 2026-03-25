# Agent P99 Latency Runbook

## Severity

Critical (P99 > 10s)

## Description

Agent {{ $labels.agent_id }} P99 latency exceeds 10 seconds. This indicates severe tail latency issues affecting the slowest requests.

## Diagnosis

1. **Identify slow requests:**

   ```bash
   ff admin agent traces {{ $labels.agent_id }} --percentile=99 --since=30m --slowest
   ```

2. **Find latency outliers:**

   ```bash
   ff admin agent latency-distribution {{ $labels.agent_id }} --since=1h
   ```

3. **Check resource constraints:**

   ```bash
   # Memory
   kubectl top pod -l agent_id={{ $labels.agent_id }}
   
   # Check for OOM kills
   kubectl get events --field-selector reason=OOMKilled | grep {{ $labels.agent_id }}
   ```

4. **Analyze slow trace:**

   ```bash
   ff admin agent trace {{ $labels.agent_id }} --id=<slowest-trace-id> --format=json
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| LLM model inference slow | Common | Use faster model; check provider |
| Cold starts | Common | Keep agent warm; optimize package |
| Memory pressure causing GC | Common | Increase memory; reduce concurrency |
| Network latency spike | Common | Use CDN; check network route |
| Database lock contention | Common | Optimize queries; use connection pool |

## Remediation Steps

### Step 1: Identify Root Cause (5 min)

```bash
# Get slow traces
ff admin agent traces {{ $labels.agent_id }} --percentile=99 --since=30m --format=json > slow_traces.json

# Analyze slow trace spans
cat slow_traces.json | jq -r '.[].spans[] | .name' | sort | uniq -c | sort -rn | head
```

### Step 2: Quick Fixes

```bash
# Warm up agent
ff admin agent warm {{ $labels.agent_id }}

# Reduce concurrency to free resources
ff admin agent scale {{ $labels.agent_id }} --concurrency=3

# Switch to faster LLM model temporarily
ff admin agent config {{ $labels.agent_id }} --set=llm.model=claude-3-haiku
```

### Step 3: Long-term Fix

- Optimize function package size
- Increase memory limits
- Use faster LLM model
- Implement caching
- Add horizontal scaling
- Optimize database queries

### Step 4: Verify Fix

```bash
# Watch P99 latency
ff admin agent metrics {{ $labels.agent_id }} --metric=latency.p99 --watch
```

## Prevention

- Set SLO: P99 < 5s
- Implement latency budget alerts
- Use latency percentiles in dashboards
- Regular performance testing
- Capacity planning
- Chaos engineering for latency
