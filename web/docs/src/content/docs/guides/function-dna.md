---
title: Function DNA
description: Learn how FunctionFly's AI-powered evolution engine analyzes and improves your functions.
sidebar:
  order: 16
---



Function DNA is an AI-powered evolution engine that analyzes your function's execution patterns, proposes code improvements, and helps your functions evolve and improve over time.

## How It Works

Function DNA monitors your functions continuously:

1. **Analysis** — Monitors execution patterns, error rates, performance bottlenecks
2. **Proposals** — AI generates potential code mutations to improve the function
3. **Review** — You review proposed changes with full diff visibility
4. **Deployment** — Accepted mutations deploy via canary with automatic rollback

---

## Enabling Function DNA

### For Individual Functions

1. Go to your function's settings
2. Click **Function DNA**
3. Enable **Auto-Evolution**
4. Configure mutation frequency

### Platform-Wide Settings

Go to **Settings → Platform → Function DNA**

| Setting | Description | Default |
|---------|-------------|---------|
| **Auto-Evolution** | Enable for all functions | Off |
| **Max Mutations/Day** | Limit proposals per function per day | 5 |
| **Default Canary %** | Initial traffic % for canary deployment | 10% |
| **Auto-Rollback** | Automatically revert on error rate spike | On |

---

## Understanding Your Function's DNA

Each function has a **DNA Profile** that represents its characteristics:

### Fitness Score

| Score | Rating | Description |
|-------|--------|-------------|
| 85-100 | Excellent | Highly optimized, minimal improvements needed |
| 65-84 | Good | Healthy, some minor optimizations possible |
| 40-64 | Fair | Moderate improvements available |
| 0-39 | Poor | Significant issues, mutation recommended |

### DNA Attributes

| Attribute | What It Measures |
|-----------|-----------------|
| **Performance** | Execution speed and resource efficiency |
| **Reliability** | Error rates and exception handling |
| **Security** | Vulnerability patterns and input validation |
| **Observability** | Logging quality and debuggability |
| **Best Practices** | Following runtime idioms and patterns |

### Viewing Your Function's DNA

```bash
# View DNA profile
ffly dna show my-function

# Output example
Function: my-function v1.2.0
Fitness Score: 78/100 (Good)

Attributes:
  Performance:     82%
  Reliability:     75%
  Security:         85%
  Observability:    70%
  Best Practices:   80%

Last Analysis: 2026-05-08T10:30:00Z
Total Mutations: 12 (8 accepted, 4 rejected)
```

---

## Mutations

### What Are Mutations?

Mutations are AI-generated code changes designed to improve specific aspects of your function.

### Types of Mutations

| Type | Purpose | Example |
|------|---------|---------|
| **Performance** | Optimize for speed/memory | Add caching, lazy loading |
| **Reliability** | Improve error handling | Add retry logic, timeouts |
| **Security** | Fix vulnerabilities | Input sanitization, escaping |
| **Cleanup** | Remove dead code | Unused imports, redundant logic |
| **Refactor** | Improve structure | Extract functions, simplify conditionals |

### Mutation Lifecycle

```
Proposed → Reviewed → Accepted/Rejected → Canary Deploy → Production
              ↑                              ↓
         (You review diff)           Auto-rollback on failure
```

### Reviewing a Mutation

When a mutation is proposed:

1. You'll receive a notification (in-app and/or email)
2. Go to **Functions → my-function → DNA**
3. Review the mutation:
   - **Summary** — What the mutation does
   - **Diff** — Side-by-side code comparison
   - **Impact** — Expected improvement (performance, reliability, etc.)
   - **Risk** — Potential risks or breaking changes

### Mutation Diff Example

```diff
- def handler(request):
-     data = fetch_from_db(request.id)
-     return {"result": data}
+ def handler(request):
+     # Add caching to reduce DB load
+     cache_key = f"user:{request.id}"
+     data = cache.get(cache_key)
+     if data is None:
+         data = fetch_from_db(request.id)
+         cache.set(cache_key, data, ttl=300)
+     return {"result": data}

+ # Impact: 60% faster for cached requests
+ # Risk: Low (additive change only)
```

### Accepting/Rejecting

| Action | What Happens |
|--------|--------------|
| **Accept** | Mutation deployed via canary (default: 10% traffic) |
| **Reject** | Mutation discarded, no action |
| **Accept & Disable Auto-Evolution** | Accept this mutation, turn off for future |
| **Report Issue** | Flag false positive to improve AI |

---

## Canary Deployments

### How Canary Works

When you accept a mutation:

1. **Initial deployment** — New code receives 10% of traffic
2. **Monitoring period** — 30 minutes (configurable)
3. **Success criteria** — Error rate stays below threshold
4. **Full rollout** — If successful, 100% traffic

### Canary Configuration

```yaml
# functionfly.jsonc
{
    "dna": {
        "canary": {
            "initial_percentage": 10,
            "increment_percentage": 25,
            "increment_interval_minutes": 10,
            "max_duration_minutes": 60,
            "rollback_on_error_rate": 0.01
        }
    }
}
```

### Manual Rollback

If issues arise during canary:

```bash
# Rollback to previous version
ffly dna rollback my-function

# Or rollback to specific version
ffly dna rollback my-function --version v1.2.0
```

---

## Auto-Evolution Settings

### Platform Settings

Go to **Settings → Platform → Function DNA**

| Setting | Description |
|---------|-------------|
| **Auto-Evolution** | Automatically propose mutations |
| **Max Mutations/Day** | Per-function daily limit |
| **Accept Threshold** | Auto-accept if confidence > threshold |
| **Notification Preferences** | Email, in-app, or both |

### Per-Function Settings

```bash
# Disable for specific function
ffly dna disable my-function

# Set custom mutation limit
ffly dna set-limit my-function --max 10

# View current settings
ffly dna settings my-function
```

---

## CLI Commands

```bash
# View DNA profile
ffly dna show <function>

# List pending mutations
ffly dna pending

# Accept a mutation
ffly dna accept <mutation-id>

# Reject a mutation
ffly dna reject <mutation-id>

# View mutation diff
ffly dna diff <mutation-id>

# Rollback
ffly dna rollback <function> [--version <version>]

# Disable/enable
ffly dna disable <function>
ffly dna enable <function>

# View history
ffly dna history <function>

# Set configuration
ffly dna set-limit <function> --max <number>
```

---

## Pricing

| Action | Cost |
|--------|------|
| DNA Analysis | Free |
| Mutation Proposals | Free |
| **Accepting a Mutation** | 50 credits |
| Rejecting a Mutation | Free |

Credits can be purchased in **Wallet** (1 USD = 100 credits).

---

## Interpreting Results

### After Accepting a Mutation

Check the impact:

```bash
ffly dna impact my-function --mutation abc123
```

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Latency p99 | 145ms | 89ms | -39% |
| Memory Usage | 128MB | 134MB | +5% |
| Error Rate | 0.5% | 0.2% | -60% |

### When to Accept vs Reject

**Accept when:**
- Performance improvement > 20%
- Risk is "Low" or "Medium"
- Change is additive (new behavior, not behavior modification)
- Error rate improvement is significant

**Reject when:**
- Breaks backward compatibility
- Introduces new dependencies
- Performance improvement < 5%
- Risk is "High"

---

## Troubleshooting

### No Mutations Proposed

**Possible causes:**
- Function is too new (needs at least 100 executions)
- DNA is disabled for the function
- Max mutations limit reached for the day

**Solutions:**
```bash
# Verify DNA is enabled
ffly dna show my-function

# Check mutation budget
ffly dna pending | grep "remaining"

# Ensure sufficient traffic
# DNA needs statistical significance (100+ invocations)
```

### Mutation Failed Canary

**What happens:**
- Error rate exceeded threshold during canary
- Automatic rollback to previous version
- Notification sent with failure details

**What to do:**
1. Review the error in logs: `ffly logs my-function --filter ERROR`
2. Check if issue was transient or persistent
3. If persistent, the mutation may need refinement
4. Contact support if rollback failed

### High Credit Usage

**To reduce costs:**
- Lower max mutations per day
- Disable auto-proposals (keep manual review)
- Batch accept/reject (same notification batch)

```bash
# Check credit usage
ffly wallet history

# Set monthly budget
ffly dna set-budget --monthly 500 --alert 80%
```

---

## Best Practices

1. **Enable on critical functions** — Prioritize functions with high traffic or business impact
2. **Review mutations promptly** — Check daily to not miss proposals
3. **Monitor after acceptance** — Watch metrics for 24 hours post-deployment
4. **Keep backups** — Maintain version history to rollback if needed
5. **Understand the diff** — Don't accept blindly; review each mutation
6. **Test in staging** — Use a staging environment for testing mutations first
