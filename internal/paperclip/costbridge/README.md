# Paperclip cost bridge

Pushes FunctionFly agent execution costs to Paperclip so Paperclip can enforce monthly budgets and auto-pause.

## Behavior

When an agent execution is recorded in the AEP handler, if the following env vars are set, the cost is also sent to Paperclip:

- `PAPERCLIP_BASE_URL` — Paperclip API base (e.g. `http://localhost:3100`)
- `PAPERCLIP_API_KEY` — Bearer token for Paperclip API
- `PAPERCLIP_COMPANY_ID` — Paperclip company UUID to attribute costs to
- `PAPERCLIP_AGENT_ID` — (optional) Paperclip agent UUID to attribute costs to; if unset, the event may still be accepted depending on Paperclip API

Cost is sent as `POST /api/companies/:companyId/cost-events` with:

- `costCents`: execution cost in cents (from FunctionFly `cost_usd * 100`)
- `occurredAt`: RFC3339 timestamp
- `agentId`: from `PAPERCLIP_AGENT_ID` if set
- `metadata`: `execution_id`, `function_uri`, and optionally `issue_id` for traceability

## Enabling budget enforcement

In Paperclip, set monthly budgets per agent (or company). When the rolled-up cost exceeds the budget, Paperclip will auto-pause the agent (no new checkouts). No changes required in FunctionFly beyond setting the env vars above.
