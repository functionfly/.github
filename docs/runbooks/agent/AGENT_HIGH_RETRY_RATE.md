# Agent High Retry Rate Runbook

## Severity

Warning (> 30% retry rate)

## Description

Agent {{ $labels.agent_id }} has a retry rate exceeding 30%, indicating transient failures or upstream instability.

## Diagnosis

1. **Check retry metrics:**

   ```bash
   ff admin agent metrics {{ $labels.agent_id }} --metric=retries --since=1h
   ```

2. **Identify failure types:**

   ```bash
   ff admin agent failures {{ $labels.agent_id }} --group-by=error_type --since=1h
   ```

3. **Check upstream dependencies:**

   ```bash
   # LLM Provider
   curl https://status.openrouter.ai/api/v2/status.json
   
   # Database
   ff admin db health
   ```

## Common Causes

| Cause | Frequency | Resolution |
|-------|-----------|------------|
| LLM provider flaky | Common | Check provider status; use fallback |
| Network instability | Common | Check network; use retry-friendly config |
| Timeout too short | Common | Increase timeout values |
| Rate limiting | Common | Implement backoff; respect limits |

## Remediation Steps

### Step 1: Identify Root Cause (5 min)

```bash
# Get retry breakdown
ff admin agent metrics {{ $labels.agent_id }} --metric=retries --format=json > retries.json

# Analyze error types
cat retries.json | jq -r '.[] | select(.outcome=="retry") | .error_code' | sort | uniq -c | sort -rn
```

### Step 2: Quick Fixes

```bash
# Increase timeout
ff admin agent config {{ $labels.agent_id }} --set=timeout.default=30s

# Enable exponential backoff
ff admin agent config {{ $labels.agent_id }} --set=retry.backoff=exponential

# Switch to more reliable LLM model
ff admin agent config {{ $labels.agent_id }} --set=llm.model=claude-3-haiku
```

### Step 3: Long-term Fix

- Use more reliable LLM provider
- Implement circuit breaker properly
- Add caching to reduce upstream calls
- Use fallback endpoints

## Prevention

- Monitor upstream provider SLAs
- Implement proper timeout and backoff
- Use multiple LLM providers
- Set up synthetic monitoring
