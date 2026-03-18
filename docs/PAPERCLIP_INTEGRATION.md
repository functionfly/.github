# Paperclip integration (control-plane for FunctionFly)

Paperclip is the internal agent control plane: goals, issues, approvals, heartbeats, and budgets. FunctionFly is the execution plane. See the plan in `.cursor/plans/` (Paperclip control-plane for FunctionFly).

## Architecture summary

- **Paperclip**: source of truth for company, org chart, goals, issues, approvals, budgets, heartbeat runs.
- **FunctionFly**: executes agent functions via `/v1/agent/execute/{author}/{name}`; enforces quota, policy, concurrency.
- **Adapter**: receives Paperclip heartbeat invocations (webhook), calls FunctionFly agent execution, posts results back as issue comments.
- **Cost bridge**: sends FunctionFly per-execution costs to Paperclip `POST /api/companies/:companyId/cost-events` for budget enforcement.

## Deployment

- [deploy/paperclip/README.md](../deploy/paperclip/README.md) — deploy Paperclip (authenticated/private) and create FunctionFly company + agents.

## Implementation

- [internal/api/handlers/paperclip/](../internal/api/handlers/paperclip/) — webhook adapter (wake + report).
- [internal/paperclip/costbridge/](../internal/paperclip/costbridge/) — cost ingestion into Paperclip.
- [docs/PAPERCLIP_GOVERNANCE.md](PAPERCLIP_GOVERNANCE.md) — issue templates and approval gates.
- [docs/PAPERCLIP_SCHEDULING_AND_COST_POLICY.md](PAPERCLIP_SCHEDULING_AND_COST_POLICY.md) — scheduling ownership and cost definition.
