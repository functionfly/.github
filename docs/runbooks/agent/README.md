# Agent Runbooks Index

This directory contains runbooks for all FunctionFly agent alerts.

## Runbook List

| Alert | Severity | Runbook |
|-------|----------|---------|
| [AgentHighErrorRate](AGENT_HIGH_ERROR_RATE.md) | Warning | High error rate (>10%) |
| [AgentCircuitOpen](AGENT_CIRCUIT_OPEN.md) | Critical | Circuit breaker OPEN |
| [AgentCircuitHalfOpen](AGENT_CIRCUIT_HALF_OPEN.md) | Warning | Circuit breaker HALF-OPEN |
| [AgentQuotaExhausted](AGENT_QUOTA_EXHAUSTED.md) | Warning | Quota >90% used |
| [AgentHighLatency](AGENT_HIGH_LATENCY.md) | Warning | P95 > 5s |
| [AgentP99Latency](AGENT_P99_LATENCY.md) | Critical | P99 > 10s |
| [AgentPolicyViolations](AGENT_POLICY_VIOLATIONS.md) | Warning | >10 violations/sec |
| [AgentQuotaViolations](AGENT_QUOTA_VIOLATIONS.md) | Warning | >5 quota violations/sec |
| [AgentHighRetryRate](AGENT_HIGH_RETRY_RATE.md) | Warning | >30% retry rate |
| [AgentConcurrencyLimitReached](AGENT_CONCURRENCY_LIMIT.md) | Warning | >90% concurrency |
| [AgentHighCost](AGENT_HIGH_COST.md) | Warning | >$10/hour |
| [AgentNoExecutions](AGENT_NO_EXECUTIONS.md) | Info | No executions in 15min |

## Pre-launch checklist

Before turning production traffic on alerts in this directory, run a **dry run**: trigger a synthetic failure in staging (or silence → unsilence a test alert), confirm on-call routing, and verify Grafana/Loki links from each runbook still resolve.

## Quick Reference

### Critical Alerts (Immediate Action Required)

1. **AgentCircuitOpen** - Check downstream services; wait for recovery or manually reset
2. **AgentP99Latency** - Identify slow operations; scale or optimize

### Warning Alerts (Investigate Within 30 min)

1. **AgentHighErrorRate** - Check logs; identify error patterns
2. **AgentQuotaExhausted** - Increase quota or optimize usage
3. **AgentHighCost** - Identify cost drivers; cap spending
4. **AgentConcurrencyLimitReached** - Scale up or rate limit

### Info Alerts (Review During Business Hours)

1. **AgentNoExecutions** - Verify agent is working; check routing

## Creating New Runbooks

When adding new agent alerts:

1. Create markdown file in this directory
2. Follow the template structure:
   - Severity level
   - Description
   - Diagnosis steps
   - Common causes table
   - Remediation steps
   - Prevention measures
3. Update this index
4. Add runbook_url to alert annotations in `deploy/monitoring/alerts/agent-alerts.yml`
