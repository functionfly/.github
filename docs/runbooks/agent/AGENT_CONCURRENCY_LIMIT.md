# Agent Concurrency Limit Runbook

## Severity

Warning (> 90% concurrency utilization)

## Description

Agent {{ $labels.agent_id }} is approaching its concurrency limit, which may cause request queuing or rejection.

## Diagnosis

1. **Check concurrency metrics:**

   ```bash
   ff admin agent metrics {{ $labels.agent_id }} --metric=concurrency --current
   ```

2. **View queued requests:**

   ```bash
   ff admin agent queue {{ $labels.agent_id }} --depth
   ```

3. **Identify concurrency bottleneck:**

   ```bash
   ff admin agent concurrency {{ $labels.agent_id }} --breakdown
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| High traffic spike | Common | Scale up; implement rate limiting |
| Slow requests blocking | Common | Optimize request handling; increase limit |
| Concurrency limit too low | Common | Increase concurrency limit |
| Agent starvation | Common | Review scheduling; adjust priorities |

## Remediation Steps

### Step 1: Quick Assessment (2 min)

```bash
# Check if requests are queuing
ff admin agent queue {{ $labels.agent_id }} --wait-time

# View concurrency trend
ff admin agent metrics {{ $labels.agent_id }} --metric=concurrency --since=30m
```

### Step 2: Immediate Actions

```bash
# Temporarily increase concurrency limit
ff admin agent scale {{ $labels.agent_id }} --concurrency=2x

# OR enable request queuing
ff admin agent config {{ $labels.agent_id }} --set=queuing.enabled=true --set=queuing.max_wait=60s
```

### Step 3: Investigate Root Cause (15 min)

```bash
# Get request breakdown by type
ff admin agent requests {{ $labels.agent_id }} --group-by=type --since=1h

# Check for slow requests
ff admin agent requests {{ $labels.agent_id }} --slowest --limit=10
```

### Step 4: Long-term Fix

- Increase concurrency limit based on capacity planning
- Optimize slow request handlers
- Implement autoscaling
- Add request coalescing

## Prevention

- Monitor concurrency trends daily
- Set up autoscaling based on concurrency
- Implement proper request timeouts
- Use connection pooling
