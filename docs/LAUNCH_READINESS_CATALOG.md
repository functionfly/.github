# FunctionFly Launch Readiness Catalog

> **Inventory of every feature shipping at GA**, organized by domain. Only items
> with code, routes, and storage present at launch are listed; deferred items
> live in [`docs/POST_LAUNCH_TODO.md`](POST_LAUNCH_TODO.md).
>
> **Last verified:** 2026-06-25 against the working tree at `/home/micro/projects/functionfly`.

**Legend**

- ✅ Ready — wired end-to-end (handler + route + storage) and reachable at GA
- ⚠️ Partial — implemented but gated, in-progress, or needs configuration
- 🔒 Enterprise — Enterprise plan only (or pay-walled behind feature flag)
- 🧪 Beta — implemented but behind an env flag (`STUDIO_ENABLED`, `GHOST_MODE_ENABLED`, `GBA_SAML_ENABLED`, `GBA_SCIM_ENABLED`, `DRE_BLOCKCHAIN_ANCHORING_ENABLED`)
- 📦 Service — lives in a sibling repo / sub-service (counts as ready when the API contract is in place)

---

## 1. Authentication & Identity (GBA — GoBetterAuth)

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **Email + password signup / login** | ✅ | HS256 JWT auth (4 h default), Argon2id hashing (OWASP time cost = 3), bcrypt legacy verification, constant-time compare. | `internal/auth/auth.go:44`, `internal/auth/password.go:15`, `internal/api/routes_auth.go:64-72` |
| **Email verification + resend** | ✅ | Token-based email verification before features unlock. | `internal/api/routes_auth.go:77-78` |
| **Password reset (request + confirm)** | ✅ | Rate-limited public reset flow. | `internal/api/routes_auth.go:99-100` |
| **Session list / revoke (single + others)** | ✅ | Multi-session management per user with revoke-other for hijack recovery. | `internal/api/routes_auth.go:128-130`, `internal/auth/session_cache.go` |
| **JWT validation & refresh** | ✅ | `/auth/refresh`, `/auth/validate`. | `internal/api/routes_auth.go:66-67,92`, `internal/auth/jwt.go` |
| **Magic link (passwordless)** | ✅ | Request + verify flows; rate-limited. | `internal/auth/magic_link.go`, `internal/api/routes_auth.go:107-110` |
| **MFA — TOTP** | ✅ | Per-user TOTP setup / verify / enable / disable. | `internal/auth/mfa.go:1` (388 LOC), `internal/api/handlers/mfa/`, `internal/api/routes_auth.go:113-118` |
| **MFA — WebAuthn / Passkeys** | ✅ | Full ceremony: register, sign, session. | `internal/auth/webauthn.go`, `internal/auth/webauthn_session.go` |
| **Backup codes** | ✅ | TOTP backup-codes (regenerate, one-time use). | `internal/auth/mfa.go` |
| **OAuth — Google & GitHub** | ✅ | OIDC + OAuth2 with tenant-scoped provider overrides. URL builders, callback handler, account linking confirm. | `internal/auth/oauth.go:143-179`, `internal/api/routes_auth.go:84-91` |
| **OAuth — per-tenant providers (BYO IdP)** | ✅ | Tenants configure their own Google / GitHub client; encrypted at rest in DB. | `internal/auth/oauth.go:189-232`, `internal/api/routes_auth.go:184-186` |
| **Captcha (Turnstile primary, recaptcha v2/v3, hCaptcha)** | ✅ | Provider-priority chain (Turnstile → reCAPTCHA v3 → v2 → hCaptcha). Required on signup, login, password reset. | `internal/captcha/auth_public.go:26`, `internal/api/middleware/turnstile.go:111`, `internal/api/routes.go:916-931` |
| **SAML SSO (Enterprise, 🧪 beta)** | ⚠️ / 🔒 | Code complete in GBA SAML plugin (`internal/auth/gba/plugins/saml/`). Routes NOT wired in `routes.go`; activated via `GBA_SAML_ENABLED=true`. Post-launch per POST_LAUNCH_TODO.md. | `internal/auth/gba/plugins/saml/handlers.go:91`, `internal/api/handlers/auth/saml.go:114-468` |
| **SCIM 2.0 provisioning (Enterprise, 🧪 beta)** | ⚠️ / 🔒 | Full SCIM 2.0 Users/Groups/Config routes; gated behind `GBA_SCIM_ENABLED=true` and `FeatureSCIM` (Enterprise). Per POST_LAUNCH_TODO.md, dashboard UI and IdP guides are post-launch. | `internal/api/routes_scim.go:15-36`, `internal/auth/scim.go`, `internal/api/handlers/auth/scim.go` |
| **RBAC / permissions** | ✅ | 17 named permissions (`PermUsersRead`, `PermBillingWrite`, …) used across admin routes + tenant resources. | `internal/auth/roles.go`, `internal/auth/constants.go`, `internal/rbac/` |
| **Account / tenant member roles** | ✅ | Owner/Admin/Member with invite + role update + revoke. | `internal/api/routes_auth.go:187-193` |
| **HIBP breach check** | ✅ | "Have I Been Pwned" lookup on signup / password change. | `internal/auth/hibp.go` |
| **Trusted device / "new device login" alert** | ✅ | Emits email on suspicious logins + remember-device token. | `internal/auth/auth.go`, `internal/email/email.go:19-21` |
| **Invite codes (invite-only launch)** | ✅ | Admin can mint / revoke signup invite codes. | `internal/api/routes_admin.go:151-153` |
| **Waitlist** | ✅ | Public waitlist signup with invite issuance. | `internal/api/routes_auth.go:74-75` |
| **Username availability + change** | ✅ | Reserved-username checker, 2/year limit, paid early-change. | `internal/api/routes_auth.go:76,144-147,195-196` |
| **User profiles (public, settings, skills, achievements, activity)** | ✅ | Public read; PATCH settings (privacy/visibility/profile/notifications/security). | `internal/api/routes_auth.go:168-179` |
| **Favorites (users ⇄ functions)** | ✅ | Add / remove / toggle / position. | `internal/api/routes_auth.go:156-161` |
| **Social follow (users + functions)** | ✅ | Followers / following / status / my-stats. | `internal/api/routes_auth.go:200-210` |

**Launch caveats**

- `JWT_SECRET` is hard-required (no fallback) — see `internal/api/routes.go:198`.
- `PRIVACY_SALT` and `GITHUB_VAULT_KEY` are also required.
- Email: Resend in production (`RESEND_API_KEY`); SMTP fallback; mock service only when `PRODUCTION_ENV != "true"` (`internal/auth/auth.go:67-76`).

---

## 2. Billing & Subscriptions

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **Plans — Free / Starter ($24/mo) / Professional ($79/mo) / Enterprise ($299/mo) / Agent Enterprise ($499/mo)** | ✅ | Plan tiers unified with Vault plans (separate Vault quota tier). | `internal/plans/limits.go:134-148`, `web/dashboard/src/lib/constants.ts` |
| **AEP (Agent Execution Plan) tiers** | ✅ | `agent_starter`, `agent_scale`, `agent_pro`, `agent_enterprise` with concurrency, calls/mo, burst ceiling, state writes/h. | `internal/plans/limits.go:142-200` |
| **Stripe Checkout & Customer Portal** | ✅ | `/billing/checkout`, `/billing/portal-session` create Stripe sessions. | `internal/api/routes_auth.go:234-242`, `internal/api/handlers/billing/` |
| **Stripe webhooks (public)** | ✅ | `POST /webhooks/stripe` is the primary event sink — subs, invoices, payments, disputes, refunds, payouts, tax. | `internal/api/routes_agent.go:29-63` (registered before registry catch-all), `internal/api/handlers/webhooks/stripe_webhook.go` |
| **Payment methods (SetupIntent, attach, default, detach)** | ✅ | Full card management via Stripe SetupIntents. | `internal/api/routes_auth.go:235-238` |
| **Subscription CRUD + cancel** | ✅ | Read current sub, cancel at period end. | `internal/api/routes_auth.go:240-241` |
| **Invoices (list, get, credit-notes, void, apply)** | ✅ | Full SOX-compliant invoice + credit-note lifecycle. | `internal/api/routes_auth.go:242`, `internal/api/routes_admin.go:286-319` |
| **Wallet / prepaid credits** | ✅ | Unified wallet per tenant + agent. Top-up, balance, freeze, adjust, reconcile. | `internal/wallet/`, `internal/api/routes_auth.go:244-245`, `internal/api/routes_admin.go:257-265` |
| **Cost allocation / metering** | ✅ | Real-time usage tracking (Redis counters), cost summary, by-function / by-period / by-region / entries. | `internal/services/realtime_usage_tracker.go`, `internal/api/routes_auth.go:326-331`, `internal/api/handlers/billing/cost_allocation_handler.go` |
| **Usage forecasting & alerts + spend caps** | ✅ | Predict next period usage; configurable alerts + per-tenant spend cap. | `internal/services/usage_forecaster.go`, `internal/services/usage_alerter.go`, `internal/api/routes_auth.go:306-323` |
| **Dunning / payment retries** | ✅ | `DunningManager` with grace-period logic; surfaced via Stripe webhook. | `internal/billing/dunning_manager.go`, `internal/api/handlers/webhooks/stripe_webhook.go` |
| **Disputes / chargebacks / refunds** | ✅ | Chargeback reconciliation, evidence upload, refund stats, open dispute views. | `internal/billing/dispute_response_manager.go`, `internal/api/routes_admin.go:299-322` |
| **Stripe Connect (creator payouts)** | ✅ | Onboarding (`/payouts/connect-account`), balance, ledger, request payout with fees, schedule preferences. | `internal/payment/payout_service.go`, `internal/api/routes.go:1323-1341` |
| **Auto-payout scheduler** | ✅ | Cron-driven payouts with approval rules. | `internal/api/routes.go:1311-1315`, `internal/scheduler/payout_scheduler.go` |
| **Affiliate / referral codes & commissions** | ✅ | User-facing affiliate endpoints + admin approval/paid marking. | `internal/api/routes_auth.go:293-297`, `internal/api/routes_admin.go:345-355` |
| **Tax exemption certificates** | ✅ | Admin review queue; VIES (EU VAT) validation. | `internal/billing/tax_exemption.go`, `internal/billing/vies_validation.go`, `internal/api/routes_admin.go:332-333` |
| **Data retention — 90 d detailed / 7 yr financial** | ✅ | `CleanupCostAllocationByRetention` + `CleanupFinancialAggregatesAfterRetention`; legal-hold aware. | `migrations/20260419131000_billing_performance_indexes.up.sql`, `internal/storage/billing/repository.go` |
| **Webhook replay & stored events** | ✅ | Replay + cleanup for failed Stripe webhooks. | `internal/api/routes_admin.go:325-329`, `internal/storage/billing_operational_repository.go` |
| **Tenant-isolated billing** | ✅ | Per-tenant Stripe accounts (`/billing/tenants/{tenant_id}/webhook`). | `internal/api/routes_auth.go:264`, `internal/api/handlers/billing/tenant_webhook.go` |
| **Pricing tiers + bundles (Founder Mode "Build Now, Pay Later")** | ✅ | Bundle catalog, founder registration, deferred billing progress, one-click deploy. | `internal/api/routes_auth.go:269-290`, `internal/billing/bundles.go` |
| **Usage exports (configurations, templates, jobs, external sync)** | ✅ | Schedule CSV / JSON / DB exports to S3, R2, GCS; external billing sync (Stripe, QuickBooks, NetSuite). | `internal/api/routes_auth.go:344-373` |
| **Pricing tiers admin CRUD** | ✅ | Create / update / delete tier. | `internal/api/routes_admin.go:274-278` |
| **MRR / ARR / churn / LTV analytics** | ✅ | Admin revenue dashboards. | `internal/api/routes_admin.go:239-251` |

**Launch caveats**

- All Stripe-backed flows require Stripe env vars (`STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `INTERNAL_WEBHOOK_SECRET`).
- Dunning retry index migration `migrations/20260419131000_billing_performance_indexes.up.sql` must be applied.

---

## 3. Function Platform / Compute

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **Function registry (CRUD, list, get, delete)** | ✅ | Public read; auth-required write. | `internal/api/routes_registry.go:130-139`, `internal/storage/registry/repository.go` |
| **Function publish / versioning / rollback / alias** | ✅ | Versions, changelogs, deploy history, alias, rollback latest. | `internal/api/routes_registry.go:142-156`, `internal/versioning/` |
| **Function execution — Go / Python / Node** | ✅ | Runtime router across local WASM pool (MicroPython + CPython-WASI) and external `wasm-pool-service`. | `internal/wasm/`, `internal/api/handlers/registry/execution/runtime_router.go`, `internal/api/routes_registry.go:99-110` |
| **WASM pool (pre-warmed MicroPython)** | ✅ | Per-tenant instance pool; warm/cold start metrics; CGO-optional. | `internal/wasmpool/`, `internal/wasmpoolservice/`, `internal/wasmpool/client/`, `internal/api/routes.go:417-445` |
| **Eager bundling (compile Python → WASM at publish time)** | ✅ | `bundler.BundleService` pre-compiles to eliminate cold-start. | `internal/bundler/`, `internal/api/routes.go:397-403` |
| **Auto-scaling / concurrency** | ✅ | Per-tenant pool + external pool split (`WASM_POOL_EXTERNAL_PERCENT`). | `internal/wasmpool/client/manager.go` |
| **Invocation API (sync, stream, replay)** | ✅ | `POST /v1/{author}/{name}[@version]`, `POST /v1/fx/...`, `POST /v1/app-run/...`, `POST /v1/run/...`. | `internal/api/routes_registry.go:99-118,131`, `internal/api/handlers/registry/handler.go` |
| **Zero-friction demo (no signup, sandboxed)** | ✅ | `GET /v1/demo`, `POST /v1/demo/execute` — public execution path for landing-page try-now. | `internal/api/routes_registry.go:84-86`, `internal/api/handlers/demo/` |
| **App playground (`/app-run/{appSlug}/{functionName}`)** | ✅ | App-scoped function execution with security middleware. | `internal/api/routes_registry.go:62-77`, `internal/api/handlers/playground/` |
| **Sandbox / secure execution coordinator** | ✅ | Per-function security: captcha, ClamAV, YARA, trust-level gating. | `internal/api/middleware/execution_coordinator.go`, `internal/api/routes.go:933-953` |
| **Logs, metrics, stats per function** | ✅ | `/functions/{a}/{n}/stats`, `/v1/functions/{id}/metrics`, `/v1/functions/{id}/logs`. | `internal/api/routes_platform.go:266-268`, `internal/api/routes_registry.go:252` |
| **Cache (memory, disk, Redis, CDN, edge)** | ✅ | Multi-tier cache with config + admin purge. | `internal/cache/cache_service.go`, `internal/api/routes_admin.go:414-418` |
| **Canary deployments** | ✅ | Create, get, update, cancel, promote, rollback, history. | `internal/api/routes_registry.go:186-192`, `internal/api/handlers/registry/canary_handler.go` |
| **Function env vars + secrets + MCP settings + embed** | ✅ | Per-function env, secrets (vault-backed), MCP toggle, embed snippet + analytics. | `internal/api/routes_registry.go:221-247` |
| **Function reviews & ratings** | ✅ | Submit review/rating; list. | `internal/api/routes_registry.go:254-257`, `routes_registry.go:79-83` |
| **Function remix (fork) with cost preview** | ✅ | Public cost preview; authenticated remix with lineage. | `internal/api/routes_registry.go:321-323` |
| **Function webhooks (per-function outgoing)** | ✅ | CRUD + deliveries + test. | `internal/api/routes_platform.go:577-583`, `internal/storage/function_webhook_repository.go` |
| **Function schedules / cron triggers** | ✅ | Create, list, manual trigger, presets. | `internal/api/routes_platform.go:276-282`, `internal/scheduler/function_scheduler.go` |
| **SDK / docs / tutorials / CDN assets** | ✅ | `/docs`, `/tutorials`, `/sdk/{sdk}/{version}/{filename}`, `/static/{category}/{path}`. | `internal/api/routes_registry.go:201-218` |
| **Deprecation + migration guides** | ✅ | Public deprecation listings + per-endpoint migration notes. | `internal/api/routes_registry.go:195-199`, `internal/api/handlers/registry/migration_handler.go` |
| **State triggers (function + webhook executors)** | ✅ | Multi-executor dispatch (internal function call + external HTTP webhook). | `internal/api/routes.go:503-535`, `internal/storage/state/trigger_engine.go` |
| **Service contracts + version negotiation** | ✅ | `/internal/contracts/*` for platform-internal inter-service contracts. | `internal/api/routes_registry.go:162-164` |
| **API versioning & deprecation (admin)** | ✅ | List, deprecate, set-default, create new versions. | `internal/api/routes_platform.go:154-159` |
| **A2A Explorer & agent runtime** | ✅ | Agent execution with tool calls and tool registry. | `internal/api/routes_agent.go:81-113`, `internal/api/routes_a2a.go` (231 LOC) |
| **Browsable & trending functions** | ✅ | `/functions/trending`, `/functions/search`. | `internal/api/routes_registry.go:318` |
| **Embed (script + analytics)** | ✅ | `GET /embed/{author}/{nameVersion}`, embed config + analytics. | `internal/api/routes_registry.go:221-225` |

**Launch caveats**

- CGO-disabled builds fall back to the external `wasm-pool-service` (separate Rust `sar` repo) — see `AGENTS.md` and `internal/wasmpool/client/manager.go`.
- `VERIFICATION_ENABLED=true` enables ClamAV + YARA trust-level checks; off by default in dev.
- Bundler pipeline must find `internal/bundler/python/micropython.wasm` or `bundler/python/micropython.wasm`; otherwise pool init is skipped.

---

## 4. Vault (Secrets)

> Zero-knowledge architecture: AES-256-GCM ciphertext + IV/salt/tag stored; the server never sees plaintext. Decryption is client-side via `web/dashboard/src/utils/vault-crypto.ts`.

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **Secrets CRUD + rotate + audit** | ✅ | List / create / get / update / delete / rotate + per-secret audit log. | `internal/api/routes_platform.go:457-466`, `internal/api/handlers/vault/`, `internal/storage/vault/repository.go` |
| **Secret versions, diff, rollback** | ✅ | Version history, diff between versions, rollback to prior version. | `internal/api/routes_platform.go:471-475` |
| **Secret dependencies graph** | ✅ | Track which functions / apps depend on a secret. | `internal/api/routes_platform.go:477-480` |
| **Dynamic credentials (Postgres targets)** | ✅ | Targets, templates, generate, revoke, lease renew. | `internal/api/routes_platform.go:505-516`, `internal/api/handlers/vault/dynamic_credentials.go` |
| **Bulk operations + export** | ✅ | Bulk delete + JSON export. | `internal/api/routes_platform.go:549-551` |
| **Quota enforcement** | ✅ | Per-tenant Redis-backed sliding window + quota (max secrets / dynamic creds / tokens / audit exports). Admin overrides via `vault_rate_limits`. | `internal/storage/vault/quota/quota.go:236-312`, `internal/api/routes.go:578-579` |
| **Vault plans (separate from platform plans)** | ✅ | Free → 25 secrets / 100 dynamic creds; Pro → 500/5 K; Team → 5 K/50 K; Enterprise → 1 M/1 M. Mapping via `platformToVaultPlan`. | `web/dashboard/src/lib/vaultPlans.ts:24-39`, `internal/plans/limits.go:36-58` |
| **MFA on vault access** | ✅ | TOTP-required before secret read/write (per-secret config). | `internal/api/routes_platform.go:483-486` |
| **IP allowlist (token-level)** | ✅ | Set IP allowlist per access token. | `internal/api/routes_platform.go:488` |
| **Expiration (per-secret)** | ✅ | Set absolute expiration + dashboard view. | `internal/api/routes_platform.go:490`, `web/dashboard/src/components/VaultEnterprise` |
| **Break-glass emergency access** | ✅ | Request / approve / deny / revoke + configurable policy. | `internal/api/routes_platform.go:492-497` |
| **Escrow (key recovery for tenants)** | ✅ | Enable / disable per tenant; status read. | `internal/api/routes_platform.go:499-501` |
| **RBAC roles + assignments** | ✅ | Roles (CRUD) + per-user assignments; "my assignments" view. | `internal/api/routes_platform.go:524-530` |
| **Secret sharing between users** | ✅ | Share / list-shared-with-me / revoke. | `internal/api/routes_platform.go:532-534` |
| **Vault SSO (SAML, team/enterprise)** | ⚠️ / 🔒 | Config endpoints exist; SAML backend beta-gated (see §1). | `internal/api/routes_platform.go:536-537`, `internal/api/handlers/vault/enterprise.go:465-477` |
| **SIEM webhooks** | ✅ | Push vault audit events to SIEM via signed webhooks. | `internal/api/routes_platform.go:539-541` |
| **Audit export** | ✅ | GET `/vault/audit/export` (CSV/JSON). | `internal/api/routes_platform.go:543` |
| **Namespaces (logical isolation)** | ✅ | List / create / delete; secrets scoped to namespace. | `internal/api/routes_platform.go:520-522` |
| **Cache stats (enterprise)** | ✅ | Per-tenant vault cache health. | `internal/api/routes_platform.go:547` |
| **Scheduled rotation** | ✅ | Cron-based rotation schedules (migration `20260624192000_vault_rotation_schedules`). | `migrations/20260624192000_vault_rotation_schedules.up.sql` |
| **SDKs (Go, Python, JS, C, Edge, Ruby, Rust, Swift, Kotlin)** | ✅ | Per-language SDKs in `sdk/`. JS + Python ship with CLI helpers. | `sdk/go`, `sdk/python`, `sdk/js`, `sdk/edge`, `sdk/go-vault-sdk`, `sdk/python-vault-sdk`, `sdk/js-vault-sdk`, `sdk/vault-secrets-operator` |

**Launch caveats**

- No server-side decrypt endpoint by design (`docs/VAULT_OPERATIONS.md`).
- Frontend crypto in `web/dashboard/src/utils/vault-crypto.ts` is the authoritative path for plaintext recovery.

---

## 5. Trust & Verification

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **DRE certificates (per execution)** | ✅ | Issue, list, get, verify certificates; passport public view. | `internal/api/routes_registry.go:266-283`, `internal/api/handlers/registry/dre/handlers.go`, `internal/dre/` |
| **DRE divergence simulation + drift reports** | ✅ | Compare executions across versions; drift timeline. | `internal/api/routes_registry.go:274-279` |
| **DRE execution timeline + by-hash lookup** | ✅ | Determinism execution graph. | `internal/api/routes_registry.go:276-278` |
| **DRE anchoring service (blockchain, 🧪 beta)** | ⚠️ | Ethereum anchoring service wired; gated behind `ANCHOR_SIGNING_KEY` + per-chain RPC + contract address. | `internal/api/routes.go:1063-1087`, `internal/dre/cert/anchoring_service.go` |
| **Trust score (function-level)** | ✅ | Per-function trust score + history + sliding-window state. | `internal/api/routes_registry.go:157-159,260`, `internal/api/routes_admin.go:439-444` |
| **Trust scheduler (hourly recalc)** | ✅ | `TRUST_SCORE_SCHEDULER_ENABLED=true` to enable. | `internal/api/routes.go:784-798`, `internal/scheduler/trust_score_scheduler.go` |
| **Trust API (external platform partners)** | ✅ | Partners, API keys (scopes: `verification:request`, `reports:submit`), trust score, batch, history, verify, report. | `internal/api/routes_trustapi.go` (208 LOC) |
| **Attestation chain + verify** | ✅ | Per-function attestation chain with verify endpoint. | `internal/api/routes_trustapi.go:115-118` |
| **Policy engine (CRUD + evaluate/batch)** | ✅ | JWT-protected write; public evaluate. | `internal/api/routes_trustapi.go:121-133` |
| **Revocation registry** | ✅ | Revoke function trust, check revoked, list, get by id. | `internal/api/routes_trustapi.go:139-144` |
| **SSE streaming for live trust scores** | ✅ | `/v1/trust/stream/sse` and `/v1/trust/stream/functions/{id}/sse`. | `internal/api/routes_trustapi.go:177-182` |
| **Trust webhooks (CRUD, deliveries, test, stats)** | ✅ | HTTP webhooks + replay support. | `internal/api/routes_trustapi.go:152-169` |
| **Trust API billing (tier pricing + partner billing)** | ✅ | Tiers, checkout, usage reports, founder mode per partner. | `internal/api/routes_trustapi.go:188-207`, `internal/api/handlers/trustapi/billing_handler.go` |
| **Expired evaluation cleanup scheduler** | ✅ | Default-on; configurable cron + max age. | `internal/api/routes.go:801-821`, `internal/scheduler/expired_evaluation_scheduler.go` |
| **Verification (approvals + signatures)** | ✅ | Request approval, decide, comment, sign function version, verify signature. | `internal/api/routes_registry.go:289-298` |

---

## 6. Agents & AI

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **AEP (Agent Execution Plan) — register, list, get, delete** | ✅ | Agents register with JWT or `X-Agent-API-Key`. | `internal/api/routes_agent.go:134-150`, `internal/api/handlers/agent/handler.go` |
| **AEP — quota / policy / usage / analytics** | ✅ | Per-agent quota, policy, executions, analytics. | `internal/api/routes_agent.go:141-148` |
| **AEP — billing summary, spend cap, cost breakdown, credit balance** | ✅ | Wallet-backed. | `internal/api/routes_agent.go:152-156` |
| **AEP — credit purchase (Stripe checkout)** | ✅ | `POST /v1/agent/{id}/credits/checkout`. | `internal/api/routes_agent.go:157` |
| **AEP — concurrency stats** | ✅ | Platform-wide concurrency telemetry. | `internal/api/routes_agent.go:160` |
| **Agent lifecycle (heartbeat, pause, resume, shutdown, terminate)** | ✅ | Full lifecycle control endpoints. | `internal/api/routes_agent.go:163-168` |
| **Root-level agent spawn (`POST /v1/agent/spawn`)** | ✅ | Standalone agent ID generation for Studio. | `internal/api/routes_agent.go:172` |
| **Agent discovery (public) & tool registry** | ✅ | `/v1/agent/discover`, `/v1/agent/tools`, `/v1/agent/tools/{name}/call`. | `internal/api/routes_agent.go:104-113` |
| **Agent function execution (`/v1/agent/execute/...`)** | ✅ | Both JWT and API-key auth; versioned. | `internal/api/routes_agent.go:107-109`, `internal/api/handlers/agentruntime/` |
| **Agent runtime (function discovery + execution + tool calls)** | ✅ | Separate agent-runtime handler with billing controller. | `internal/api/routes_agent.go:90-101` |
| **Agent wallets** | ✅ | Per-agent wallet balance, transactions, low-balance alerts (`AGENT_WALLET_LOW_BALANCE_USD`, default $5). | `internal/api/handlers/agent/wallet_handler.go`, `internal/wallet/` |
| **Agent team memory (auto-injection into prompts)** | ✅ | Team-memory middleware auto-enabled on `/agent/generate` and `/agent/execute`. | `internal/api/routes_agent.go:117-128`, `internal/team_memory/` |
| **SEBG (Self-Evolving Backend Graph)** | ✅ | Proposals, decide, tier, evolve, ROI endpoints. | `internal/api/routes_agent.go:181`, `internal/api/handlers/agent/sebg_handler.go` |
| **Evolution API** | ✅ | Suggestions, auto-enable, history. | `internal/api/routes_agent.go:185`, `internal/api/handlers/agent/evolution_handler.go` |
| **Daemon API (always-on agents)** | ✅ | Start / stop / status / config. | `internal/api/routes_agent.go:189`, `internal/api/handlers/agent/daemon_handler.go` |
| **Swarm / marketplace / economy** | ✅ | Swarm services (platform controller, metrics, workers, unfair advantage engine). | `internal/api/routes_agent.go:177`, `internal/agent/swarm/`, `internal/api/routes_admin.go:460-468` |
| **Marketplace & economy services** | ✅ | Marketplace, financial transactions, economy. | `internal/agent/marketplace/`, `internal/agent/economy/` |
| **Unfair Advantage Engine (internal)** | ✅ | RDLab, stealth pipeline, internal function generation. | `internal/api/routes_admin.go:460-468`, `internal/api/handlers/agent/unfair_advantage_handler.go` |
| **Conversations (executable, with bounties + reactions + attachments)** | ✅ | Full conversation service: list/create/messages/bounties/reactions/attachments/read receipts. | `internal/api/routes_agent.go:232-263`, `internal/api/handlers/conversations/` |
| **Conversations WebSocket** | ✅ | Real-time messaging. | `internal/api/routes_agent.go:227` |
| **Browser automation (Playwright)** | ✅ | `internal/api/handlers/browser/`, registered when `s.browserSvc != nil`. | `internal/api/routes_agent.go:192-198` |
| **AI Composer / Gallery** | ✅ | `/v1/ai/composer/generate[/stream]`, `/refine[/stream]`; health + status. | `internal/api/routes.go:1351-1359`, `internal/api/ai_proxy_handler.go` |
| **Recommendations engine** | ✅ | Personalized function recommendations, composable search, triple search. | `internal/api/routes_registry.go:167-173`, `internal/api/handlers/recommendations/` |
| **Categorization (taxonomy, tags, AI categorization)** | ✅ | Public read; AI analyze endpoint; per-function tag management. | `internal/api/routes_platform.go:188-198`, `internal/api/handlers/categorization/` |
| **Factory (function generation pipeline)** | ✅ | Opportunity discovery (GitHub, Reddit, SO, Google), approve / reject, pipeline run, A/B experiments. | `internal/api/routes_platform.go:162-185`, `internal/agent/factory/` |
| **Factory pipeline scheduler** | ✅ | Cron-driven discovery + generation. | `internal/api/routes.go:714-724`, `internal/scheduler/factory_pipeline_scheduler.go` |
| **AI Composer (FRG-backed function generation)** | ✅ | `POST /frg/functions/generate` and `POST /frg/compose`. | `internal/api/routes_frg.go:167-168,294-295` |
| **🧠 FlyMind AI service (📦, separate Python service)** | ✅ | OpenAI / Anthropic / DeepInfra / Fireworks / Groq / MiMo / Ollama / OpenRouter / StepFun / Together providers. Connectors to orchestrator via Atlas. Used by support, chat, team memory, DNA, factory, AI composer. | `ai-service/src/providers/`, `ai-service/src/main.py`, `internal/support/ai_client.go`, `internal/team_memory/auto_updater.go` |
| **Azure AI integrations (📦 via FlyMind service)** | ⚠️ | Azure AI Search, Speech, OpenAI, Document Intelligence integrations live in the FlyMind ai-service (`ai-service/src/integrations/`); orchestrator side hooks via `atlas.py` + `atlas_grpc.py`. | `ai-service/src/integrations/atlas.py`, `ai-service/src/integrations/atlas_grpc.py` |
| **AI Gateway (`ai-gateway`, 📦)** | ✅ | Standalone gateway in `ai-gateway/` subproject; orchestrator calls via `AI_SERVICE_URL`. | `ai-gateway/`, `internal/support/ai_client.go`, `internal/dna/service.go:102-110` |
| **Chat (real-time AI chat with connector integration)** | ✅ | `chat/ws` WebSocket, connector registry, AI service client. | `internal/api/routes_platform.go:596-599`, `internal/api/handlers/chat/` |
| **Brain (long-term memory store with RAG)** | ✅ | Connectors + brain CRUD; queries via `/v1/brain/*`. | `internal/api/routes_platform.go` (registerConnectorRoutes/registerBrainRoutes), `internal/api/handlers/brain/`, `internal/api/handlers/connectors/` |

**Launch caveats**

- `ai-service` is a separate deployable; orchestrator requires `AI_SERVICE_URL`, `AI_SERVICE_API_KEY`. If unset, fallbacks to built-in rule-based assistant (`internal/support/ai_client.go:321`).
- Agent team memory middleware activates only for `/agent/generate` and `/agent/execute` paths.
- Browser service is only registered when `s.browserSvc != nil`.

---

## 7. HyperFrames / Studio

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **FRG (Function Registry + Live Runtime Graph)** | ✅ | Versioned, composable function graphs with streaming execution, NATS or in-memory event bus, trigger router. Public read; auth+rate-limit write. | `internal/api/routes_frg.go` (311 LOC), `internal/frg/` |
| **FRG graph CRUD + publish + remix** | ✅ | Author/scoped graph CRUD, publish versions, fork. | `internal/api/routes_frg.go:139-154,272-283` |
| **FRG graph execution (`/gx/{author}/{name}@{version}`)** | ✅ | Streaming graph execution with FXCERTs (DRE-signed when configured). | `internal/api/routes_frg.go:158-159,286-287` |
| **FRG AI composition** | ✅ | `POST /frg/compose` + `POST /frg/functions/generate`. | `internal/api/routes_frg.go:167-168,294-295` |
| **FRG webhooks (dynamic + fixed)** | ✅ | `POST /webhook/{path}`, `POST /api/webhooks/graph/{graph_id}`. | `internal/api/routes_frg.go:172-174,298-299` |
| **FRG instance management (status/stop/resume)** | ✅ | Long-running instance lifecycle. | `internal/api/routes_frg.go:162-164,290-291` |
| **Backend-as-a-Graph: auto-generated REST/GraphQL APIs** | ✅ | `autoGenAPIHandler.RouteRegistrar` mounts REST routes for published graphs. | `internal/api/routes_frg.go:177-188,301-310`, `internal/frg/api/autogen_handler.go` |
| **FRG semantic discovery (vector search)** | ✅ | `/frg/discover` (public) + `/frg/graphs/{a}/{n}/optimizations`. | `internal/api/routes_frg.go:144-145,275-276`, `internal/api/handlers/frg/` |
| **Studio (🧪 beta)** | ⚠️ | All `/v1/studio/*` routes gated behind `STUDIO_ENABLED=true`. Includes collaboration events/activity, tasks, extensions, settings, code editor (format/save/undo/redo/history), DevOps pipelines / environments / regions. Per POST_LAUNCH_TODO.md, global flag → per-tenant flag post-launch. | `internal/api/routes_platform.go:291-348`, `internal/api/handlers/studio/` |
| **GSAP animation skill (frontend)** | ✅ | Reference skill for `gsap.to / from / fromTo`, timelines, stagger, performance. | `.agents/skills/gsap/SKILL.md` (used by frontend code generation) |
| **Remotion video rendering (📦, via inference.sh)** | ✅ | Reference skill + integration for rendering `useCurrentFrame`, `useVideoConfig`, `spring`, `interpolate`, `AbsoluteFill`, `Sequence` to MP4. | `.agents/skills/remotion-render/SKILL.md` |
| **AI Composer (UI in dashboard)** | ✅ | Frontend at `web/dashboard/src/pages/AIComposerPage`. | `web/dashboard/src/pages/AIComposerPage/` |

---

## 8. Rankings & Gamification

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **City Rankings™ — leaderboard + categories + movers + map** | ⚠️ | Handlers + scorer + cache + storage + cron job fully implemented; **routes NOT registered at launch** (per POST_LAUNCH_TODO.md). Data pipeline runs on `internal/jobs/cityranking/recompute.go`. | `internal/api/handlers/cityranking/handler.go`, `internal/api/routes_cityranking.go` (defined, not called), `internal/storage/cityranking/`, `internal/jobs/cityranking/recompute.go`, `web/dashboard/src/pages/CityRankingsPage` |
| **City Rankings™ — "my city" + opt-out + IP-resolve** | ⚠️ | Same status — code complete, route registration deferred. | `internal/api/routes_cityranking.go:33-37` |
| **City Rankings™ — builders / ambassador / universities per metro** | ⚠️ | Same status. | `internal/api/routes_cityranking.go:42-44` |
| **City Ambassadors program** | ⚠️ | Ambassador model + sync job implemented; surfaced via city-ranking routes. | `internal/jobs/cityambassador/`, `internal/storage/cityranking/repository.go`, `web/dashboard/src/pages/AmbassadorsPage` |
| **City Wars™ (quarterly bracket)** | ⚠️ | Handler + routes defined, but not registered at launch. Dashboard UI present (`CityWarsAdminPage`). | `internal/api/routes_citywar.go` (17 LOC), `web/dashboard/src/pages/CityWarsAdminPage` |
| **University Rankings™** | ⚠️ | Handler + storage + seed CSV implemented (`universities_seed.csv` ~170 US universities); **routes not registered**. | `internal/api/routes_universityranking.go` (29 LOC), `internal/api/handlers/universityranking/handler.go`, `internal/storage/universityranking/`, `internal/jobs/universityranking/`, `web/dashboard/src/pages/UniversityRankingsPage` |
| **Company Rankings** | 📝 | Specified in `docs/CITY_RANKINGS.md` as future; listed as post-launch TODO. | `docs/POST_LAUNCH_TODO.md:9` |

**Launch caveats**

- The `registerCityRankingRoutes`, `registerUniversityRankingRoutes`, and `registerCityWarRoutes` helpers exist but are **not called** from `internal/api/routes.go`. Treating City/University Rankings as "implemented but not shipped" until route registration is added.
- Cron-driven scoring continues to populate `data/` regardless of route status.

---

## 9. Admin & Platform Management

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **Admin auth + IP allowlist + session + CSRF + rate limit** | ✅ | Admin `/v1/admin/*` subrouter enforces IP allowlist → admin session → advanced rate limit → CSRF. | `internal/api/routes_admin.go:78-123` |
| **Admin MFA enforcement + force-disable** | ✅ | `admin/mfa/force-disable` for support interventions. | `internal/api/routes_admin.go:389`, `internal/api/middleware/` |
| **Tenant CRUD** | ✅ | List / get / create / update / delete (HMAC-signed). | `internal/api/routes_admin.go:135-139` |
| **User management (list, stats, invite, get, update, delete)** | ✅ | HMAC-signed writes. | `internal/api/routes_admin.go:142-148` |
| **Signup invite codes (list / create / revoke)** | ✅ | Per-tenant invite-only launch flow. | `internal/api/routes_admin.go:151-153` |
| **Audit events (legacy + new)** | ✅ | Both `audit-events` (platform audit) and `audit` (admin audit). | `internal/api/routes_admin.go:156,168-169` |
| **IP allowlist management** | ✅ | List / create / update / delete / toggle; self-check access. | `internal/api/routes_admin.go:159-165` |
| **Security events + alerts** | ✅ | List, stats, review; alert CRUD with HMAC signing. | `internal/api/routes_admin.go:172-183` |
| **Maintenance mode (global + per-tenant + templates + schedule + audit)** | ✅ | Full maintenance lifecycle. | `internal/api/routes_admin.go:186-199`, `internal/storage/maintenance_repository.go` |
| **Platform backends + providers admin** | ✅ | Enable/disable providers; backend visibility toggle. | `internal/api/routes_admin.go:201-212` |
| **Incident management** | ✅ | Create / list / update / resolve. | `internal/api/routes_admin.go:215-219` |
| **System health / metrics / edge status** | ✅ | `/admin/health`, `/admin/status/edge`, `/admin/system/metrics`. | `internal/api/routes_admin.go:222-224` |
| **Admin dashboard endpoints (activity / revenue / quick-stats)** | ✅ | Used by `web/admin-dashboard` SPA. | `internal/api/routes_admin.go:227-229` |
| **Analytics (platform + tenant + MRR/ARR/churn/LTV)** | ✅ | Full analytics surface. | `internal/api/routes_admin.go:232-251` |
| **Billing summary, wallets (freeze / unfreeze / close / adjust / reconcile)** | ✅ | Comprehensive wallet ops. | `internal/api/routes_admin.go:255-265` |
| **Payout approvals + rules** | ✅ | Admin payout approval queue + rule CRUD. | `internal/api/routes_admin.go:268-272` |
| **Pricing tiers, subscriptions, invoices admin CRUD** | ✅ | HMAC-signed. | `internal/api/routes_admin.go:274-289` |
| **Usage admin (per-tenant + metrics)** | ✅ | Real-time + state-fabric usage. | `internal/api/routes_admin.go:358-363` |
| **Cost allocation admin (tenant + chargeback report)** | ✅ | Internal chargeback accounting. | `internal/api/routes_admin.go:366-367` |
| **Disputes / refunds / credit notes / chargebacks** | ✅ | Full dispute + refund ops + SOX-compliant credit notes. | `internal/api/routes_admin.go:299-322`, `internal/api/handlers/admin/disputes.go` |
| **Billing webhook replay & cleanup** | ✅ | Stored webhook inspector + replay + retention cleanup. | `internal/api/routes_admin.go:325-329` |
| **Tax exemption review queue** | ✅ | Admin review with VIES validation. | `internal/api/routes_admin.go:332-333` |
| **Feedback admin (list / stats / analytics / export / status update)** | ✅ | HMAC-signed status updates. | `internal/api/routes_admin.go:336-340` |
| **Affiliate code admin (codes / referrals / commissions / approve / paid)** | ✅ | Full referral program admin. | `internal/api/routes_admin.go:345-355` |
| **Cache admin (stats / purge-all / per-function / per-version)** | ✅ | HMAC-signed purges. | `internal/api/routes_admin.go:414-418` |
| **Retention settings + cleanup stats + manual cleanup** | ✅ | 90 d / 7 yr retention policies; legal-hold aware. | `internal/api/routes_admin.go:421-426` |
| **Cloudflare analytics admin view** | ✅ | Aggregated Cloudflare metrics. | `internal/api/routes_admin.go:429` |
| **Oversight dashboard (trust / execution audit / fraud / economic leaderboard / block / investigate)** | ✅ | Internal trust + abuse-monitoring surface. | `internal/api/routes_admin.go:432-437`, `internal/api/handlers/admin/oversight.go` |
| **Admin function CRUD (cross-tenant)** | ✅ | List/get/update/delete/toggle + deployments/logs/metrics. | `internal/api/routes_admin.go:392-399` |
| **Admin registry (visibility / pricing / backfill READMEs / DRE bootstrap regen)** | ✅ | Cross-tenant registry control. | `internal/api/routes_admin.go:402-412` |
| **State Fabric admin (stats / settings / suspend / resume / cleanup / TTL stats)** | ✅ | Operational control plane. | `internal/api/routes_admin.go:471-478` |
| **Trigger engine stats + queue stats** | ✅ | Observability for state-fabric trigger engine. | `internal/api/routes_admin.go:482-499` |
| **State / consciousness / DNA schedulers (admin access)** | ✅ | Cron lifecycles with admin observability. | `internal/api/routes.go:1247-1302` |
| **Unfair Advantage Engine (admin-only)** | ✅ | RDLab, stealth pipeline, internal function generation. | `internal/api/routes_admin.go:460-468` |
| **Newsletter admin** | ✅ | Admin newsletter composer / send. | `internal/api/routes_admin.go:1196`, `internal/api/handlers/newsletter/` |
| **Enterprise audit log export (HMAC-signed)** | ✅ | Audit log list/filters/export with audit signing key. | `internal/api/routes.go:489-497`, `internal/api/handlers/enterprise/` |
| **Admin SPA (`web/admin-dashboard`)** | ✅ | Separate React SPA with its own IP-allowlisted login. 30+ admin pages (Auth, Users, Billing, Backends, Cache, Compliance, Email, Factory, Features, Feedback, Functions, IP Allowlist, Incidents, Maintenance, Monitoring, Newsletter, Providers, Retention, Security Events, Sessions, Tax Exemptions, etc.). | `web/admin-dashboard/src/pages/*.tsx` |

---

## 10. Developer Experience

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **CLI (`ff`, 📦, separate repo)** | ✅ | Referenced from orchestrator and dashboard; Fly login flows use `/auth/login` over `/v1` or bare. | `cli/ffly/` (this repo has placeholder; lives in `functionfly/cli` per AGENTS.md) |
| **Python SDK (`flypy`)** | ✅ | Full SDK with examples; ships `python-interpreter`, `flypy/` package, agent SDK integration tests. | `sdk/python/flypy/`, `sdk/python/examples/`, `sdk/python/test_*.py` |
| **JS SDK (monorepo)** | ✅ | Multi-package JS SDK in `sdk/js/packages/`. | `sdk/js/packages/` |
| **Edge SDK** | ✅ | For Cloudflare Workers / Edge runtimes. | `sdk/edge/` |
| **Vault SDKs (Go, Python, JS)** | ✅ | Vault client SDKs for each runtime. | `sdk/go-vault-sdk/`, `sdk/python-vault-sdk/`, `sdk/js-vault-sdk/` |
| **Vault Secrets Operator** | ✅ | Kubernetes operator for vault secret sync. | `sdk/vault-secrets-operator/` |
| **Additional SDKs (C, Ruby, Rust, Swift, Kotlin, GitHub Actions)** | ✅ | Each ships in its own subdirectory. | `sdk/c/`, `sdk/ruby/`, `sdk/rust/`, `sdk/swift/`, `sdk/kotlin/`, `sdk/github-actions/` |
| **API docs (OpenAPI 3 + Swagger UI)** | ✅ | `/swagger`, `/swagger/doc.json`, `/swagger/doc.yaml` (auth-gated in production). | `internal/api/routes.go:1373-1381`, `internal/api/docs/` |
| **Internal tutorials (`/tutorials/*`)** | ✅ | Getting-started, API usage, function development, interactive examples. | `internal/api/routes_registry.go:208-213`, `internal/api/handlers/registry/tutorials_handler.go` |
| **SDK / docs CDN (`/sdk/{sdk}/{version}/{filename}`, `/static/{category}/{path}`)** | ✅ | Versioned static asset serving. | `internal/api/routes_registry.go:215-218` |
| **Examples (`/examples`)** | ✅ | 13 example projects: `ai-sentiment`, `email-notification`, `external-api`, `file-storage`, `kv-counter`, `python`, `python-microvm`, `rust`, `stdlib-publish`, `typescript`, `typescript-wasm`, `wasm`, `webhook-notifier`. | `examples/` |
| **Webhooks (function + trust + billing)** | ✅ | Function webhooks (`/v1/function-webhooks/*`), trust webhooks (`/v1/webhooks/*`), Stripe webhooks (`/webhooks/stripe`), Paperclip (`/v1/paperclip/*`), tenant-isolated billing (`/v1/billing/tenants/{id}/webhook`). | `internal/api/routes_platform.go:577-583`, `internal/api/routes_trustapi.go:149-169`, `internal/api/routes_agent.go:36-63` |
| **Public docs site (`web/docs`, Astro)** | ✅ | Astro Starlight-powered user docs at port 4322. | `web/docs/` (60+ MDX pages: `getting-started`, `quick-start`, `functions`, `registry`, `bundles`, `sdks`, `cli`, `security`, `pricing`, etc.) |
| **Marketing site (`web/site`, Astro)** | ✅ | Marketing site at port 4321: index, pricing, plans, security, compliance, SLA, trust, ambassadors, university rankings, city wars, partnerships, careers, contact, blog, legal (terms/privacy/dpa). | `web/site/src/pages/*.astro` |
| **Trust site (`web/site/trust.astro`)** | ✅ | Standalone trust page surfaced on marketing site. | `web/site/src/pages/trust.astro` |
| **Demo / try-now API (`/v1/demo/*`)** | ✅ | Public, no-auth function execution for landing-page demo. | `internal/api/routes_registry.go:84-86` |

---

## 11. Operations / Infrastructure

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **Health endpoints** | ✅ | `/health`, `/healthz`, `/health/detailed`, `/health/check`, `/health/dna`. | `internal/api/routes.go:1365-1369`, `internal/health/` |
| **Prometheus metrics** | ✅ | `/metrics` (auth-gated in prod) + per-route HTTP metrics middleware. | `internal/api/routes.go:1007,1370`, `internal/monitoring/` |
| **Status WebSocket** | ✅ | `/ws/v1/status` real-time status push. | `internal/api/routes.go:1371`, `internal/api/handlers/status/handler.go` |
| **`.well-known/functionfly.json`** | ✅ | Discovery manifest for partner integrations. | `internal/api/routes.go:1363` |
| **PostgreSQL + pgvector** | ✅ | Primary store; pgvector for embeddings (FRG, support, recommendations). | `docs/LOCAL_POSTGRES_17.md`, `migrations/` (820+ files) |
| **Redis** | ✅ | Rate limits, real-time usage counters, presence, pub/sub, vault quota, notification pool. | `internal/cache/`, `internal/wallet/`, `internal/services/realtime_usage_tracker.go`, `internal/api/routes.go:219-225` |
| **Cloudflare — R2 / Workers / Tunnel / Pages / DNS** | ✅ | DNS managed at `deploy/dns/`; Tunnel + Workers + Pages configured in `docs/CLOUDFLARE.md`. | `docs/CLOUDFLARE.md`, `deploy/dns/`, `internal/api/handlers/admin/cloudflare.go` |
| **Caddy edge** | ✅ | Edge config in `deploy/edge/`, `edge-targets/`, `deploy-edge.sh`. | `deploy/edge/`, `edge-targets/` |
| **Migration system** | ✅ | 820+ timestamped migrations (`YYYYMMDDHHMMSS_description.sql`); helper `scripts/create-migration.sh`, validator `scripts/validate-migrations.sh`. Idempotent SQL via `IF NOT EXISTS`. | `migrations/`, `MIGRATIONS.md`, `scripts/` |
| **Deployment scripts** | ✅ | `Makefile`, `deploy/` scripts, `docker-compose.{yml,dev,local,production,staging,monitoring,admin,auth,runtime}.yml`, Fly.io configs (`fly.toml`, `fly.staging.toml`, etc.). | `deploy/`, `Makefile`, `fly*.toml`, `docker-compose*.yml` |
| **Backups & disaster recovery** | ✅ | Documented runbooks; backup storage cost comparison. | `docs/DISASTER_RECOVERY.md`, `docs/DISASTER_RECOVERY_RUNBOOK.md`, `docs/BACKUP_STORAGE_COST_COMPARISON.md`, `docs/RUNBOOK.md` |
| **Tenant DB runbook** | ✅ | Tenant-isolated DB operations. | `docs/TENANT_DB_RUNBOOK.md` |
| **Maintenance mode** | ✅ | Global middleware checks maintenance flag (admin-controlled). | `internal/api/middleware/maintenance.go`, `internal/api/routes_admin.go:186-199` |
| **Per-tenant feature middleware** | ✅ | Feature flag middleware (`internal/api/middleware/feature.go`) used for gated features. | `internal/api/routes_timemachine.go`, `internal/api/routes_consciousness.go` |
| **Environment middleware** | ✅ | Per-request env header parsing. | `internal/api/middleware/environment.go` |
| **Body size limits** | ✅ | 1 MB default, configurable per-route. | `internal/api/routes.go:964-966` |
| **CSRF middleware (Upstash-backed)** | ✅ | Tokens issued per session; required for mutating requests. | `internal/api/middleware/csrf.go`, `internal/api/routes.go:1029` |
| **Rate limiters (auth, wallet, vault, provider, public, MFA, admin)** | ✅ | One per domain. | `internal/api/middleware/` |
| **Notification pool (pgxpool LISTEN)** | ✅ | Postgres NOTIFY/LISTEN for real-time WS push. | `internal/api/routes.go:218-225` |
| **Tracing middleware** | ✅ | Per-request trace spans. | `internal/api/routes.go:961` |

---

## 12. Marketing & Docs Sites

| Site | Status | Stack | Pages / Routes | Location |
|---|---|---|---|---|
| **Marketing site** | ✅ | Astro (port 4321) | index, pricing, plans, registry, bundles, security, compliance, sla, trust, ambassadors, universities, city-wars, company-rankings, partnerships, careers, contact, about, dpa, terms, privacy, vulnerability, mcp, agent-execution, state-fabric, for-agents, changelog, blog. | `web/site/src/pages/*.astro` (30+ pages) |
| **Public docs site** | ✅ | Astro Starlight (port 4322) | `getting-started`, `quick-start`, `functions`, `registry`, `sdks`, `cli`, `security`, `bundles`, `pricing`, `providers`, `statefabric`, `runtimes`, `secrets-vault`, `api-reference`, `trust-api`, `trust-and-verification`, `trust-protocol-spec`, `agents`, `analytics`, `migration-from-competitors`, `open-source-strategy`, `roadmap`, `function-webhooks`, `deploy-keys`, `deployment`. | `web/docs/src/content/docs/**` (60+ MDX pages) |
| **Trust site section** | ✅ | Embedded in `web/site/trust.astro` | Single landing page. | `web/site/src/pages/trust.astro` |
| **Monthly State Reports** | ✅ | Astro content collection | `web/site/src/content/reports/2026-{03,04,05,06}.md`. | `web/site/src/content/reports/` |
| **Dashboard (SPA, port 3000)** | ✅ | Vite + React (port 3000) | 143 page folders (Billing hub, Vault, Vault Enterprise, Agents, Agent Wallets, Admin Factory, City/University Rankings, Time Machine, State Fabric, Brain, Connectors, Chat, Conversations, Decisions, DevOps, DNA, Vault, Triggers, etc.). | `web/dashboard/src/pages/` |
| **Admin dashboard SPA** | ✅ | Vite + React (separate Vite build) | 30+ admin pages (Audit, Auth Audit, Backends, Billing, Blog, Cache, Changelog, City Reviews, Cloudflare Analytics, Content, Content Calendar, Dashboard, Economic Leaderboard, Email, Employees, Execution Audit, Factory, Features, Feedback, Fraud Detection, Function Detail, Functions, IP Allowlist, Incidents, Login, Maintenance, Monitoring, Newsletter, Providers, etc.). | `web/admin-dashboard/src/pages/` |

---

## 13. Cross-Cutting Features

| Feature | Status | Description | Key code paths |
|---|---|---|---|
| **Rate limiting (per-domain)** | ✅ | Auth, wallet, vault, provider, public, MFA, admin, message, agent. | `internal/api/middleware/` |
| **Email — Resend (prod) / SMTP (fallback) / Mock (dev)** | ✅ | `RESEND_API_KEY` required in `PRODUCTION_ENV=true`. | `internal/email/resend.go`, `internal/email/fromenv.go`, `internal/auth/auth.go:67-76` |
| **Transactional emails (signup, magic link, password reset, new device, security alert, low wallet, waitlist, newsletter)** | ✅ | 10+ transactional templates. | `internal/email/email.go:15-27` |
| **In-app notifications (Postgres LISTEN + WS push)** | ✅ | Per-user; unread count; mark-as-read; preferences. | `internal/api/routes_auth.go:333-340`, `internal/notification/`, `internal/api/routes.go:218-225` |
| **Internationalization (i18n)** | ✅ | Frontend i18n in `web/dashboard/src/lib/i18n`. Translation TODO at `TRANSLATION_TODO.md`. | `web/dashboard/src/lib/i18n/`, `TRANSLATION_TODO.md` |
| **GDPR — Privacy service (export, delete, consent)** | ✅ | `PrivacyService` + handlers; `/v1/privacy/*` routes. | `internal/api/routes_privacy.go` (64 LOC), `internal/privacy/`, `internal/api/handlers/privacy/` |
| **Audit triggers (Postgres)** | ✅ | `audit_trigger_function()` with `::inet` cast for `ip_address`. | `migrations/` (audit_trigger_function) |
| **Login attempt repository** | ✅ | Brute-force defense via `LoginAttemptRepository`. | `internal/api/routes.go:168` |
| **Cookie consent** | ✅ | Frontend `cookie-consent/` component. | `web/dashboard/src/components/cookie-consent/` |
| **Multi-tenancy** | ✅ | Per-tenant data isolation enforced in repositories + middleware. | `internal/storage/`, `internal/api/middleware/auth.go` |
| **Webhook signature verification** | ✅ | HMAC signature middleware for sensitive endpoints (admin write paths). | `internal/api/middleware/advanced_security/hmac.go` |
| **Production security middleware (DDoS, geo-blocking, rate limit, input validation)** | ✅ | Activated only when `PRODUCTION_ENV=true`. | `internal/api/routes.go:977-997`, `internal/api/middleware/advanced_security/` |
| **Compliance certifications** | ✅ | SOC 2 / ISO / HIPAA messaging surfaces. Compliance endpoints in admin. | `internal/api/routes_platform.go:128,380-385`, `web/site/src/pages/compliance.astro`, `web/site/src/pages/sla.astro` |
| **Privacy policy / Terms / DPA** | ✅ | Public marketing pages. | `web/site/src/pages/{privacy,terms,dpa}.astro` |

---

## Cross-Domain Highlights

### FRG (Function Registry + Live Runtime Graph) — the strategic differentiator
- Streaming graph execution with NATS or in-memory event bus
- DRE-signed execution certificates (`FXCERTs`) when keys configured
- AI composition (`/frg/compose`) and function generation (`/frg/functions/generate`)
- Auto-generated REST/GraphQL APIs from published graphs (`autoGenAPIHandler.RouteRegistrar`)
- Public semantic discovery at `/frg/discover`
- See: `internal/frg/`, `internal/api/routes_frg.go`

### DRE (Decentralized Registry Endpoint)
- Per-execution certificates, drift reports, divergence simulation, execution timeline
- Ethereum anchoring service (`internal/dre/cert/anchoring_service.go`) is wired but **gated** by env config
- See: `internal/api/routes_registry.go:266-283`, `docs/TRUST_PROTOCOL_SPEC.md`

### State (stateful function memory)
- Per-tenant state with optional encryption, snapshots, time-travel, permissions
- Triggers (function-call or HTTP webhook executors) with cron support
- See: `internal/api/routes_platform.go:387-412`, `internal/storage/state/repository.go`

### Time Machine (per-plan replay windows)
- 24 h / 72 h / 30 d / 90 d / unlimited replay windows by plan
- See: `internal/plans/limits.go:80-106`, `internal/api/routes_timemachine.go`

### Function Consciousness
- Per-tenant awareness score + auto-fix proposals (Pro/Enterprise/Agent Enterprise tiers)
- See: `internal/api/routes_consciousness.go`, `internal/scheduler/consciousness_scheduler.go`

### Plugin Manager + Marketplace
- Plugin CRUD + sandboxing + permissions + telemetry + analytics
- See: `internal/api/routes_platform.go:351-369`, `internal/api/handlers/plugin/`, `internal/api/handlers/marketplace/`

### Team Memory (Shared Brain)
- Team-scoped memories, semantic search, extraction approval workflow
- See: `internal/api/routes_platform.go:222-232`, `internal/team_memory/`

---

## Launch Readiness Summary

| Status | Count | Notes |
|---|---|---|
| ✅ Ready | ~140 features | Shipped at GA |
| ⚠️ Partial (gated / needs config) | ~12 features | Studio, Ghost Mode, SAML SSO, SCIM, DRE blockchain anchoring, City/University Rankings routes (code complete, route registration deferred per POST_LAUNCH_TODO.md) |
| 🔒 Enterprise-only | ~20 features | SSO, SCIM, advanced security, audit export, dedicated pools, MFA-required vaults, micro-VMs, etc. |
| 🧪 Beta (env-gated) | 4 features | Studio (`STUDIO_ENABLED`), Ghost Mode (`GHOST_MODE_ENABLED`), SAML (`GBA_SAML_ENABLED`), SCIM (`GBA_SCIM_ENABLED`) |
| 📦 External service | 3 services | FlyMind AI service (`ai-service/`), CLI (`cli/ffly`, separate repo), SAR runtime (`sar` Rust repo, external — referenced as `wasm-pool-service`) |

### Hard prerequisites before first GA deploy
1. **Env vars required at startup** (orchestrator-api will refuse to start without them):
   - `JWT_SECRET` (≥ 32 bytes) — `internal/api/routes.go:198`
   - `PRIVACY_SALT` — `internal/api/routes.go:385`
   - `GITHUB_VAULT_KEY` — `internal/api/routes.go:243`
   - `RESEND_API_KEY` **or** SMTP config when `PRODUCTION_ENV=true` — `internal/auth/auth.go:67-76`
2. **PostgreSQL 17** with pgvector running on `5432` — see `docs/LOCAL_POSTGRES_17.md`
3. **Redis** running on `6379`
4. **Migrations applied** (use `--skip-migrations` only in dev due to duplicate sequence history)
5. **Optional but recommended for full GA**: `OPENROUTER_API_KEY`, `STRIPE_*`, `PRODUCTION_ENV=true`, `CLAMAV_URL`, `YARA_URL`, `PROMETHEUS_URL`, `ANCHOR_*` (for DRE blockchain)

### Deferred / explicitly post-launch (see `docs/POST_LAUNCH_TODO.md`)
- City Rankings™ UI polish, City Wars™ bracket, Ambassador program expansion
- University Rankings™ internationalization, University Wars, ambassador program
- Studio: per-tenant feature flag, frontend nav badges
- Ghost Mode: per-tenant flag, nav badges
- DRE blockchain anchoring: HSM integration, per-tenant flag
- SAML SSO: production enablement, IdP config UI, SSO audit log dashboard
- SCIM: IdP integration guides (Okta/Azure AD/OneLogin), dashboard UI, token rotation UI, audit logging, E2E sandbox tests
- Magic link / passwordless for enterprise (optional)
- Webhook notifications for SCIM user/group changes
- Outbound SCIM notifications to IdP (optional)
- JIT provisioning from SAML assertions
- Phone support UI (`EnterpriseSupportPage`)
- City/University Ranking route registration (code complete, intentionally not wired at launch)
