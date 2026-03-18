# Paperclip governance: issue templates and approval gates

Use Paperclip issues as the single queue for FunctionFly engineering and ops. Use approvals for high–blast-radius actions.

## Issue templates (definition of done)

Create these as Paperclip issue templates or checklists so agents and humans know what “done” means.

### Engineering (feature / bug)

- [ ] Scope agreed (title + short description)
- [ ] Branch/PR created and linked
- [ ] CI green
- [ ] Review completed (or waived with reason)
- [ ] Merged to main (or target branch)
- [ ] Release / deploy step noted if needed

### Infra / ops (deploy, config change)

- [ ] Change described and rollback plan noted
- [ ] Staging verified (if applicable)
- [ ] Approval obtained (see approval gates below)
- [ ] Runbook or doc updated if behavior changes

### Security (vuln triage, patch)

- [ ] Severity and impact noted
- [ ] Patch or mitigation identified
- [ ] PR or change linked
- [ ] Approval if production-facing

## Approval gates (required before execution)

These actions **must** have a Paperclip approval (board approve/reject) before an agent or human runs them.

| Action | Approval type | Notes |
|--------|----------------|--------|
| **Production deploy** | `deploy_production` or custom | Any deploy that affects production (including canary). |
| **Database migration** | `db_migration` | Schema or data migrations on shared DB. |
| **Secrets rotation** | `secrets_rotation` | Rotating API keys, certs, or vault secrets. |
| **Auth / policy change** | `auth_policy` | Changing auth config, RBAC, or security policy. |
| **Budget or billing change** | `budget` | Changing agent/company budgets or billing. |

### How to enforce

1. In Paperclip, create approval types (or use existing ones) that match the table above.
2. For each high-risk action, create an issue and link an approval request (or create the approval from the issue).
3. Do **not** run the action until the approval is approved in Paperclip.
4. Optionally: in runbooks or automation, require a “check approval” step (e.g. link to the approval or issue) before executing.

## First two weeks: run work through Paperclip

- **Week 1**: Stand up Paperclip, create FunctionFly company and agents, seed ~10 real issues (4 eng, 3 ops, 2 security, 1 docs). Require **issue checkout** before any agent starts work. Add the approval types above.
- **Week 2**: Implement the webhook adapter and cost bridge. Run all new eng/ops work through Paperclip; measure duplicated work rate, cycle time, and budget burn accuracy.

## See also

- [PAPERCLIP_INTEGRATION.md](PAPERCLIP_INTEGRATION.md) — architecture and deployment
- [PAPERCLIP_SCHEDULING_AND_COST_POLICY.md](PAPERCLIP_SCHEDULING_AND_COST_POLICY.md) — scheduling ownership and cost definition
