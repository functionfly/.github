# Function Version Management — Security, API & UI

> **Purpose:** Runbook for registry/function versioning after the production-readiness hardening (2026-05).  
> **Broader spec:** See [VERSIONING_SYSTEM.md](./VERSIONING_SYSTEM.md) for lifecycle concepts and examples.

---

## Two layers (do not confuse)

| Layer | Storage | Primary API | Who uses it |
|-------|---------|-------------|-------------|
| **Registry publish** | `registry_function_versions` | `POST /functions/publish` (auth) | Studio, CLI, publish flow — creates/updates version **rows** + WASM/source |
| **Registry lifecycle** | Same table + `version_state`, aliases, rollbacks | `/v1/functions/{functionId}/versions/*` | Dashboard owner UI, automation — publish/archive/deprecate/alias/rollback **state** |

Publishing code goes through **registry publish**. Promoting a draft to `published`, setting `latest`/`stable`, rollbacks, and deprecation use the **platform** routes (by `functionId` UUID).

---

## Security model (launch-critical)

### Authorization helpers

Code: `internal/api/handlers/registry/authz.go`, `internal/api/handlers/version/authz.go`

| Check | Rule |
|-------|------|
| **Owner** | `owner_user_id ==` authenticated user, or role `admin` / `super_admin` |
| **View public** | `visibility` is `public`, `unlisted`, or empty — anyone |
| **View private** | Owner, same `tenant_id` as function, or platform admin |
| **Publish author** | `req.Author` must match user `username` or `email` (case-insensitive). Namespace `functionfly` is **admin-only** for new publishes |

### What requires ownership (403 if not owner)

- `POST /functions/publish` when the function **already exists**
- All mutating `/v1/functions/{functionId}/versions/*` (publish, archive, deprecate, alias, changelog)
- `POST /v1/functions/{functionId}/rollback` and `.../versions/{version}/rollback`
- Registry canary: `POST/PATCH/DELETE /functions/{author}/{name}/canary*` (handler checks owner)

### What requires view access (403 on private for strangers)

- `GET /v1/functions/{functionId}/versions` (list)
- `GET /v1/functions/{functionId}/versions/{version}`
- `GET /functions/{author}/{name}/source` — **full source** gated by visibility
- `GET .../versions/compare`, `.../lineage`

### Owner-only reads (auth required)

- `GET /v1/functions/{functionId}/rollbacks`
- `GET .../versions/{version}/deployments` (+ single deployment must belong to that function)

### Platform API version admin (not function versions)

Routes under `/api/versions` — **not** per-function:

- Mutations require permission `system.write` (or admin/super_admin role).
- Do **not** use plain `RequireAuth` only; any authenticated user could otherwise deprecate `v1`/`v2`.

### Internal service contracts

Routes: `/internal/contracts`, `/internal/contracts/{service}`, `/internal/contracts/negotiate`

- Protected by `RequireInternalSecret` middleware.
- **Production:** set `INTERNAL_WEBHOOK_SECRET` and send header `X-Internal-Webhook-Secret` on every request.
- **Local dev:** if secret unset, only loopback (`localhost` / `127.0.0.1`) is allowed.

---

## Rollback behavior

- `POST /v1/functions/{functionId}/rollback` — rolls back to **previous** published version.
- `POST /v1/functions/{functionId}/versions/{version}/rollback` — rolls back **to** `{version}` (path wins unless body sets `toVersion`).
- Implementation sets `latest` alias to the target version and records a row in `rollback_records`.
- Strategies are accepted in the API; handler effectively performs **immediate** alias swap today.

---

## Dashboard UI (where things live)

| Surface | Path / file | Notes |
|---------|-------------|-------|
| **Owner version management** | `web/dashboard/src/pages/FunctionPage/FunctionVersionsSection.tsx` | Shown when API returns `is_owner: true` on function profile |
| **Canary** | `web/dashboard/src/pages/FunctionPage/FunctionCanarySection.tsx` | Same owner gate; uses `/functions/{author}/{name}/canary` |
| **API client** | `web/dashboard/src/api/versions.ts` | `versionsApi`, `registryCanaryApi` |
| **Public version list** | Function page — read-only cards for non-owners | `GET /functions/{author}/{name}/versions` |
| **FRG graph versions** | `web/dashboard/src/components/frg/panels/VersionSelector.tsx` | Graph versions; compare uses platform compare when linked registry function exists |
| **Deploy picker** | `web/dashboard/src/pages/RegistryDeployPage` | Version select for deploy only |
| **Docs site** | `web/docs/.../VersionSelector.tsx` | Public version dropdown only |

Owner detection: `GET /v1/functions/{author}/{name}` includes `is_owner` and `visibility` when the caller is the owner (`EnrichFunctionInfoForViewer` in registry query handler). Private functions hide source in the UI unless `is_owner`.

---

## Environment & ops

| Variable | Purpose |
|----------|---------|
| `INTERNAL_WEBHOOK_SECRET` | Required in production for `/internal/contracts/*` |
| Header `X-Internal-Webhook-Secret` | Must match the env var for internal contract calls |

---

## Key code locations (for agents)

| Area | Location |
|------|----------|
| Platform version HTTP handlers | `internal/api/handlers/version/handler.go` |
| Route registration | `internal/api/routes_registry.go` (function versions), `internal/api/routes_platform.go` (API versions) |
| Version handler wiring | `internal/api/routes.go` — `versionhandler.NewHandler(versionRepo, registryRepo)` |
| Registry publish + ownership | `internal/api/handlers/registry/publish.go` |
| Source + list versions (public read) | `internal/api/handlers/registry/query.go` |
| Canary (ownership in handler) | `internal/api/handlers/registry/canary.go` |
| Internal secret middleware | `internal/api/middleware/internal_secret.go` |
| DB migration (version_state, changelog, etc.) | `migrations/20260309000000_versioning_system.up.sql` |

---

## Pre-launch checklist

- [ ] Confirm `INTERNAL_WEBHOOK_SECRET` is set in production and contract callers send the header.
- [ ] Smoke-test: non-owner cannot `POST` publish to another user's `author/name`.
- [ ] Smoke-test: non-owner with `functionId` UUID gets 403 on platform publish/rollback/deprecate.
- [ ] Smoke-test: private function source returns 403 for anonymous/non-owner.
- [ ] Smoke-test: owner sees Versions + Canary on function page; publish/rollback/deprecate work.
- [ ] Confirm only `system.write` / admin can mutate `/api/versions`.

---

## Related docs

- [VERSIONING_SYSTEM.md](./VERSIONING_SYSTEM.md) — full versioning spec and API examples
- [AGENTS.md](../AGENTS.md) — local dev and codebase map
