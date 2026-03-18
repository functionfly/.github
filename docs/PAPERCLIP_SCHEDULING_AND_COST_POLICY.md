# Paperclip scheduling and cost policy

This document locks in two decisions that make or break the Paperclip × FunctionFly integration.

## 1. Scheduling ownership

- **Paperclip** owns **work governance**:
  - What to do next (backlog, priorities)
  - Task checkout (atomic “who is working on this”)
  - Approvals (deploy, migration, secrets, policy)
  - Budgets and cost rollups
  - Audit trail (issues, comments, cost events)

- **FunctionFly** owns **execution constraints**:
  - Quota (rate limits, spend caps per agent)
  - Policy (behavioral rules, depth limits)
  - Concurrency (slots per agent/plan)
  - Runtime (WASM/VM execution)

We do **not** run two schedulers in conflict. Paperclip decides *what* and *when* (at a business level); FunctionFly decides *whether* a given execution is allowed and runs it.

## 2. Cost definition

- **Initial policy**: One **unified “cents”** number per execution.
  - FunctionFly reports `cost_usd` per execution (today often 0; placeholder for future token/compute pricing).
  - The cost bridge converts to integer cents and sends to Paperclip `POST /api/companies/:companyId/cost-events`.
  - Paperclip rolls up by agent/company and enforces monthly budgets; no separate token vs compute breakdown at first.

- **Later (optional)**: Split into token vs compute vs other in Paperclip or in our reporting, once the single-number flow is stable and we have real cost data.

## Summary

| Topic | Owner | Notes |
|-------|--------|--------|
| What to do next | Paperclip | Goals, projects, issues, checkout |
| Approvals | Paperclip | Deploy, migration, secrets, policy |
| Budgets | Paperclip | Monthly caps, auto-pause |
| Execution allowed? | FunctionFly | Quota, policy, concurrency |
| Cost per execution | FunctionFly | Single `cost_usd` → Paperclip cents |
