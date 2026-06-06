# Execution Receipt — Operations Guide

The Execution Receipt is a public, shareable URL generated after every
successful function execution. This doc covers:

- What the feature does (recap)
- Environment variables
- How to enable / disable
- How to monitor in production
- Common operational questions

See `plans/EXECUTION_RECEIPT.md` for the full architecture plan and
`plans/EXECUTION_RECEIPT.md §14` for the day-1 / day-7 rollout.

---

## What it is

Every successful function execution creates (or updates) a row in
`registry_executions_public` with:

- A public nanoid (`public_id`) used in the URL: `/r/:public_id`
- The function's name, author, runtime, version
- The recorded input + output
- Duration, cached/live flag, verification status

The receipt page (`/r/:id`) is a React SPA page that:
- Renders the function header, stats, input/output JSON, and input/output
  schema
- Lets any visitor re-run the function (subject to visibility + paid gate)
- Has a one-click "Fork this function" CTA that drops unauthenticated
  users into the signup flow with a `next=` back to the editor
- Shows a subtle "Powered by FunctionFly" badge (the distribution lever)

The backend also runs a "milestone" worker that fires in-app
notifications + emails when a function's run count crosses a configured
threshold (default: 1, 10, 100, 1000, 10000). The notification includes
a pre-filled tweet-intent URL — we never auto-tweet, the owner always
clicks to confirm.

---

## Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `RECEIPT_ENABLED` | `true` | Master kill switch. When `false`, no routes are registered and no hooks fire. |
| `RECEIPT_AUTO_GENERATE` | `true` | When `true`, every successful execution creates a receipt. Set to `false` to revert to the legacy behaviour (only `SideEffects == "none"` creates receipts). |
| `RECEIPT_PUBLIC_BASE_URL` | `https://functionfly.com/r` | Base URL for the public receipt pages. Used in the `share.url` and tweet text. |
| `RECEIPT_OG_BASE_URL` | `https://functionfly.com/og/receipt` | Base URL for the dynamic OG image endpoint. |
| `RECEIPT_TWITTER_HANDLE` | `functionfly` | Handle used in tweet-intent URLs (`via=` and `@mention`). |
| `RECEIPT_SIGNING_KEY` | (empty) | Optional HMAC key for signing share URLs. Empty disables signing. |
| `RECEIPT_MILESTONE_ENABLED` | `false` | Master switch for the milestone worker. **Default `false` so the first rollout doesn't spam owners with notifications.** |
| `RECEIPT_MILESTONE_THRESHOLDS` | `1,10,100,1000,10000` | Comma-separated list of run counts that trigger a milestone event. |
| `RECEIPT_MILESTONE_CHANNELS` | `inapp,email` | Channels to fire on each milestone. `tweet_intent` is also recognised (it's a no-op fan-out — the URL is computed on the read path). |
| `RECEIPT_SWEEP_ENABLED` | `false` | Enables the daily sweep scheduler that back-fills missed milestones. |
| `RECEIPT_SWEEP_CRON` | `0 3 * * *` | Cron for the sweep (UTC). 3am by default to spread load with the other daily schedulers. |

---

## Enabling in production

Recommended order (mirrors the rollout plan §9):

1. **Week 1 day 1:** Apply the migration. Keep `RECEIPT_ENABLED=false`.
   Verify the migration runs cleanly on a staging DB.
2. **Week 1 day 2:** Set `RECEIPT_ENABLED=true` and
   `RECEIPT_AUTO_GENERATE=true`. Monitor for 24h. Check the Grafana
   panel for the new metrics (see below).
3. **Week 1 day 4:** Enable `RECEIPT_MILESTONE_ENABLED=true` for an
   internal test function only (create a test function with a known
   owner, then run it 11 times to cross the 10 threshold).
4. **Week 1 day 7:** Flip `RECEIPT_MILESTONE_ENABLED=true` globally
   and `RECEIPT_SWEEP_ENABLED=true`.

---

## Monitoring

### Metrics (Prometheus)

| Metric | Labels | Meaning |
|---|---|---|
| `receipt_get_total` | `cache`, `status` | `/v1/receipts/:id` reads. `cache=l1` is a Redis hit, `cache=db` is a miss. |
| `receipt_run_total` | `status` | `/v1/receipts/:id/run` invocations, by HTTP status. |
| `receipt_fork_total` | `status` | `/v1/receipts/:id/fork-payload` reads. |
| `receipt_view_total` | `status` | `/v1/receipts/:id/view` analytics pings. |
| `receipt_trending_total` | (none) | `/v1/receipts/trending` reads. |
| `receipt_milestone_fired_total` | `threshold`, `channel` | Milestone events fired. |
| `receipt_milestone_duplicates_total` | (none) | Milestone events skipped because the dedupe key was already present. Should grow slowly. |

### Grafana panel (recommended)

Add a "Receipt funnel" panel:

```
receipt_get_total{status="200"}         # views
  → receipt_run_total{status="200"}     # runs
    → receipt_fork_total{status="200"}  # forks
      → signups attributed via ?r= cookie
```

The last hop requires a small frontend follow-up to set the cookie
when the user lands from a receipt — out of scope for week 1.

### Logs

Every successful execution emits a structured log line:

```json
{
  "level": "info",
  "msg": "receipt backfilled",
  "public_id": "V1StGXR8_Z5jHi3B-myT",
  "function_id": "..."
}
```

Errors are logged at `warn` level (e.g. backfill failure) with enough
context to debug. They never fail the user-facing request because the
backfill runs in a goroutine.

### Alerts (suggested)

- Page on `receipt_run_total{status="5xx"} / receipt_run_total > 0.01`
  for 5m.
- Page on `rate(receipt_get_total{status="5xx"}[5m]) > 1` for 5m.

---

## Common questions

### How do I revoke a receipt?

Owner-only `POST /v1/receipts/:id/revoke` with a body of
`{ "reason": "..." }`. The receipt row gets a `revoked_at` timestamp
and the public read returns 410 Gone. The audit row goes into
`receipt_revocations`.

The UI lives at `FunctionPage` (the per-function page in the dashboard)
under the receipts tab — that work is out of scope for week 1 but the
endpoint is ready.

### Can a malicious actor spam the public `/run` endpoint?

The endpoint is gated:
- `can_run` must be `true` (function must be `visibility = "public"` and
  `price_per_call = 0`)
- IP rate-limited to 10/min via Redis sliding window
- Behind the same execution verification gate as the normal `/v1/fx/...`
  endpoint

For paid functions, the public run returns 402 Payment Required. The
owner can also revoke the receipt to remove the entry point entirely.

### How does the privacy gate work?

The receipt package uses the existing `PrivacyService` (added in an
earlier feature) to re-sanitize input/output on the read path. The same
gate is also applied on the write path inside `generateExecutionID`. So
PII never makes it onto a receipt page even if a row was created before
a policy change.

### How does this affect the existing `/replay` page?

The existing `/replay/:execId` page continues to work as a dashboard
deep link. Receipts are the *upgrade* of replays: richer metadata, the
run panel, the fork CTA, OG tags. We can leave both routes live and
gradually shift internal links to `/r/:id`.

### Can I disable milestones per-user?

Yes — a per-user opt-out is in the migration as
`users.receipt_milestones_enabled`. The worker reads it before
inserting the in-app notification. UI to set this flag is out of
scope for week 1.

---

## Rollback

If a critical issue is discovered:

1. Set `RECEIPT_ENABLED=false` and restart the orchestrator. Public
   routes immediately 404; the existing replay page keeps working.
2. To roll back the DB schema, run
   `migrations/20260601180000_execution_receipt.down.sql`. This drops
   the new columns and the new tables; existing rows keep their
   receipts but lose the denormalized metadata. The `shareable` /
   `public_id` / `input_json` / `output_json` columns are unchanged
   so the legacy replay feature continues to work.
